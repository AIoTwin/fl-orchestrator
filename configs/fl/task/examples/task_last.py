#
# Cifar10 - IID
# lr = 0.1, batch_size = 32, epochs = 3
#
from collections import defaultdict, OrderedDict
import numpy as np
import math
from typing import Dict, List
import torch
import torch.nn as nn
import time
import torch.nn.functional as F
from flwr_datasets import FederatedDataset
from flwr_datasets.partitioner import IidPartitioner
from torch.utils.data import DataLoader
from torchvision.transforms import Compose, Normalize, ToTensor
from datetime import datetime
from collections import defaultdict, Counter
import random
from copy import deepcopy

class Net(nn.Module):
    def __init__(self):
        super(Net, self).__init__()

        self.conv1 = nn.Conv2d(3, 32, kernel_size=3, padding=1)
        self.bn1 = nn.BatchNorm2d(32)

        self.conv2 = nn.Conv2d(32, 64, kernel_size=3, padding=1)
        self.bn2 = nn.BatchNorm2d(64)

        self.pool = nn.MaxPool2d(2, 2)
        self.dropout = nn.Dropout(0.25)

        self.conv3 = nn.Conv2d(64, 128, kernel_size=3, padding=1)
        self.bn3 = nn.BatchNorm2d(128)

        self.fc1 = nn.Linear(128 * 4 * 4, 256)
        self.fc2 = nn.Linear(256, 10)

    def forward(self, x):
        x = self.pool(F.relu(self.bn1(self.conv1(x))))  # 32x32 -> 16x16
        x = self.pool(F.relu(self.bn2(self.conv2(x))))  # 16x16 -> 8x8
        x = self.pool(F.relu(self.bn3(self.conv3(x))))  # 8x8 -> 4x4

        x = x.view(-1, 128 * 4 * 4)
        x = self.dropout(F.relu(self.fc1(x)))
        x = self.fc2(x)
        return x

def get_weights(net):
    return [val.cpu().numpy() for _, val in net.state_dict().items()]


def set_weights(net, parameters):
    params_dict = zip(net.state_dict().keys(), parameters)
    state_dict = OrderedDict({k: torch.tensor(v) for k, v in params_dict})
    net.load_state_dict(state_dict, strict=True)


fds = None  # Cache FederatedDataset
BIG_CLIENT_IDS = {0, 5, 13, 22} #21, 23}

class_counts = None
"""
##non iid
def load_data(
    partition_id: int,
    num_partitions: int,
    batch_size: int,
    small_frac: float = 1.00,
    big_frac: float = 0.2,
    seed: int = 42,  # kept for signature compatibility; unused
):
    print("partition id:", partition_id)

    global fds
    global class_counts
    if fds is None:
        partitioner = IidPartitioner(num_partitions=num_partitions)
        fds = FederatedDataset(
            dataset="uoft-cs/cifar10",
            partitioners={"train": partitioner},
        )

    ds = fds.load_partition(partition_id)

    frac = big_frac if partition_id in BIG_CLIENT_IDS else small_frac
    if not (0.0 < frac <= 1.0):
        raise ValueError(f"Invalid fraction {frac}. Must be in (0, 1].")

    # --- pure slice, no shuffle ---
    subset_size = max(1, min(len(ds), int(round(len(ds) * frac))))
    ds = ds.select(range(subset_size))

    class_counts = count_labels_as_list(ds, num_classes=10)
    print(f"Class counts...: {class_counts}")
    
    # --- 80/20 split, no shuffle ---
    if subset_size == 1:
        train_ds, test_ds = ds, ds.select([])
    else:
        split = ds.train_test_split(test_size=0.2, shuffle=False)
        train_ds, test_ds = split["train"], split["test"]

    # --- transforms + loaders ---
    tfm = Compose([ToTensor(), Normalize((0.5, 0.5, 0.5), (0.5, 0.5, 0.5))])

    def apply_transforms(batch):
        batch["img"] = [tfm(img) for img in batch["img"]]
        return batch

    trainset = train_ds.with_transform(apply_transforms)
    testset  = test_ds.with_transform(apply_transforms)

    print("batch_size:", batch_size)
    trainloader = DataLoader(trainset, batch_size=batch_size, shuffle=True)
    testloader  = DataLoader(testset,  batch_size=batch_size)

    return trainloader, testloader
"""


# -------------------- HARD-CODED DISTRIBUTION --------------------
STATIC_DISTRIB: Dict[str, List[int]] = {
    "n18": [191, 0,   0,   0,   0, 206, 0, 200, 0,   0],
    "n23": [195, 198, 0,   0,   0, 0,   206, 0,   0,   0],
    "n21": [0,   0,   224, 219, 0, 0,   0,   0,   0,   198],
    "n24": [0,   0,   0,   224, 206, 0, 204, 0,   0,   0],
    "n19": [0,   209, 209, 0,   0, 0,   0,   0,   0,   216],
    "n20": [0,   210, 0,   0,   0, 198, 0,   0,   228, 0],
    "n12": [0,   0,   0,   209, 203, 0, 0,   199, 0,   0],
    "n11": [0,   0,   0,   0,   0, 216, 198, 0,   0,   216],
    "n9":  [249, 178, 0,   0,   189, 0, 0,   0,   0,   0],
    "n10": [0,   0,   211, 0,   0, 198, 0,   215, 0,   0],
    "n22": [0,   0,   210, 0,   0, 0,   190, 217, 0,   0],
    "n13": [208, 0,   0,   0,   0, 0,   232, 0,   0,   225],
    "n5":  [221, 0,   0,   0,   212, 0, 0,   0,   224, 0],
    "n26": [0,   218, 0,   0,   0, 0,   0,   206, 238, 0],
    "n8":  [210, 0,   238, 0,   0, 0,   198, 0,   0,   0],
    "n17": [0,   0,   0,   229, 187, 0, 0,   228, 0,   0],
    "n25": [220, 0,   0,   0,   0, 0,   0,   0,   214, 213],
    "n6":  [0,   0,   0,   219, 212, 0, 199, 0,   0,   0],
    "n27": [0,   0,   0,   190, 0, 210, 0,   215, 0,   0],
    "n7":  [222, 0,   0,   0,   0, 195, 0,   0,   0,   208],
    "n30": [0,   208, 0,   0,   0, 199, 0,   0,   221, 0],
    "n4":  [0,   0,   202, 181, 0, 221, 0,   0,   0,   0],
    "n29": [0,   0,   214, 206, 0, 0,   0,   0,   0,   195],
    "n28": [212, 0,   0,   223, 0, 0,   0,   0,   224, 0],
}

# n4..n13 -> pid=n-4 (0..9), n17..n30 -> pid=n-7 (10..23)
def client_name_from_partition_id(pid: int) -> str:
    n = pid + 4 if 0 <= pid <= 9 else pid + 7
    return f"n{n}"

# -------------------- GLOBALS --------------------
_BASE = None                      # full CIFAR-10 train split via Flower
_PLAN: Dict[int, List[int]] = {}  # pid -> exact indices into _BASE

def _count_labels_as_list(dataset, num_classes=10):
    c = Counter(dataset["label"])
    return [c.get(i, 0) for i in range(num_classes)]

def _build_static_plan() -> None:
    """Build one deterministic, non-overlapping plan of CIFAR-10 indices per pid.
    Uses ONLY Flower's FederatedDataset to load the data."""
    global _BASE, _PLAN
    if _PLAN:
        return

    # Load the entire train split through Flower (no direct HF call)
    fds_full = FederatedDataset(
        dataset="uoft-cs/cifar10",
        partitioners={"train": IidPartitioner(num_partitions=1)},
    )
    ds = fds_full.load_partition(0)  # full train split (50k)

    labels = ds["label"]
    pool = {c: [i for i, y in enumerate(labels) if y == c] for c in range(10)}

    needed = [client_name_from_partition_id(pid) for pid in range(len(STATIC_DISTRIB))]
    if set(needed) != set(STATIC_DISTRIB.keys()):
        missing = set(needed) - set(STATIC_DISTRIB.keys())
        extra   = set(STATIC_DISTRIB.keys()) - set(needed)
        raise ValueError(f"STATIC_DISTRIB mismatch. Missing={sorted(missing)} Extra={sorted(extra)}")

    plan: Dict[int, List[int]] = {}
    for pid in range(len(STATIC_DISTRIB)):  # 0..23
        cname = client_name_from_partition_id(pid)
        target = STATIC_DISTRIB[cname]
        chosen: List[int] = []
        for cls, k in enumerate(target):
            if k == 0:
                continue
            if len(pool[cls]) < k:
                raise ValueError(f"Not enough class {cls} samples for {cname}: need {k}, have {len(pool[cls])}")
            chosen.extend(pool[cls][:k])
            pool[cls] = pool[cls][k:]
        plan[pid] = sorted(chosen)

    _BASE = ds
    _PLAN = plan

# -------------------- PUBLIC LOADER --------------------
def load_data(
    partition_id: int,
    num_partitions: int,
    batch_size: int,
    small_frac: float = 1.00,  # ignored for static plan
    big_frac: float = 0.2,     # ignored for static plan
    seed: int = 42,            # kept for signature compatibility
):
    print("partition id:", partition_id)

    # ---- FAST PATH: if num_partitions == 1, return the WHOLE dataset (Flower) ----
    if num_partitions == 1:
        fds_all = FederatedDataset(
            dataset="uoft-cs/cifar10",
            partitioners={"train": IidPartitioner(num_partitions=1)},
        )
        ds = fds_all.load_partition(0)  # full 50k train split
        class_counts = _count_labels_as_list(ds, num_classes=10)
        print(f"class counts: {class_counts} (total={len(ds)})")

        # Deterministic 80/20 split, no shuffle
        n = len(ds)
        split_at = int(round(n * 0.8))
        train_ds = ds.select(range(split_at))
        test_ds  = ds.select(range(split_at, n))

    else:
        # ---- STATIC PLAN (deterministic images per client) ----
        _build_static_plan()
        if partition_id in _PLAN:
            sel_idx = _PLAN[partition_id]
            ds = _BASE.select(sel_idx)
            class_counts = _count_labels_as_list(ds, num_classes=10)
            print(f"class counts: {class_counts} (total={len(ds)})")

            n = len(ds)
            split_at = int(round(n * 0.8))
            train_ds = ds.select(range(split_at))
            test_ds  = ds.select(range(split_at, n))
        else:
            # Fallback: standard IID partition via Flower if pid not in static plan
            fds = FederatedDataset(
                dataset="uoft-cs/cifar10",
                partitioners={"train": IidPartitioner(num_partitions=num_partitions)},
            )
            ds = fds.load_partition(partition_id)
            class_counts = _count_labels_as_list(ds, num_classes=10)
            print(f"class counts (IID fallback): {class_counts} (total={len(ds)})")

            # Deterministic 80/20 split
            n = len(ds)
            split_at = int(round(n * 0.8))
            train_ds = ds.select(range(split_at))
            test_ds  = ds.select(range(split_at, n))

    # ---- transforms + loaders (same as before) ----
    tfm = Compose([ToTensor(), Normalize((0.5, 0.5, 0.5),
                                         (0.5, 0.5, 0.5))])

    def apply_transforms(batch):
        batch["img"] = [tfm(img) for img in batch["img"]]
        return batch

    trainset = train_ds.with_transform(apply_transforms)
    testset  = test_ds.with_transform(apply_transforms)

    print("batch_size:", batch_size)
    trainloader = DataLoader(trainset, batch_size=batch_size, shuffle=True)
    testloader  = DataLoader(testset,  batch_size=batch_size, shuffle=False)
    return trainloader, testloader
# NON-IID setup
"""
def load_data(
    partition_id: int,
    num_partitions: int,
    batch_size: int,
    small_frac: float = 1.0,
    big_frac: float = 1.0,
    seed: int = 42,
    small_classes_per_client: int = 3,
):

    #If num_partitions == 1: give the single client the entire dataset (no class filtering, no subsampling).
    #Otherwise:
    #  - Big clients (in BIG_CLIENT_IDS): see ONLY classes {7,8,9}, subsampled by big_frac.
    #  - Small clients (not in BIG_CLIENT_IDS): see random classes from {0..6}, subsampled by small_frac.
   

    batch_size = 32
    print("Hello proggramer, here is a non-iid dataset")
    global fds
    global class_counts
    if fds is None:
        partitioner = IidPartitioner(num_partitions=num_partitions)
        fds = FederatedDataset(
            dataset="uoft-cs/cifar10",
            partitioners={"train": partitioner},
        )
    performance_improvement = False
    # --- Special case: single partition => whole dataset, no filtering/subsampling ---
    if num_partitions == 1:
        ds = fds.load_partition(0)  # full train split as a single partition
        frac = 1.0                  # ensure we keep everything below
        print("Single-partition mode: using full dataset.")
    else:
        # Load this client's IID partition
        ds = fds.load_partition(partition_id)
        print("Hello 1", datetime.now().strftime("%H:%M"))

        LAST_THREE = {0, 1, 2, 3, 4, 5 ,6 , 7, 8, 9}

        if (partition_id in BIG_CLIENT_IDS) and performance_improvement:
            # --- BIG CLIENTS: only last three classes ---
            allowed_classes = LAST_THREE

            def _filter(batch):
                return [lbl in allowed_classes for lbl in batch["label"]]

            ds = ds.filter(_filter, batched=True)
            frac = big_frac
            print("Hello 2", datetime.now().strftime("%H:%M"))
        else:
            # --- SMALL CLIENTS: random classes from {0..6} ---
            #######pool = list(set(range(10)) - LAST_THREE)  # {0..6}
            LAST_THREE = {7, 8, 9}
            #pool = list(set(range(10)) - LAST_THREE) #PERFORMANCE IMPROVEMENT
            pool = list(set(range(10))) #PERFORMANCE DEGRADATION 

            if small_classes_per_client > len(pool):
                raise ValueError(
                    f"small_classes_per_client={small_classes_per_client} "
                    f"exceeds available non-last-three classes ({len(pool)})."
                )
            rng = random.Random(f"{seed}-{partition_id}")  # deterministic per client
            chosen = rng.sample(pool, k=small_classes_per_client)
            allowed_classes = set(chosen)
            
            print("Allowed classes: ", allowed_classes)

            def _filter(batch):
                return [lbl in allowed_classes for lbl in batch["label"]]

            ds = ds.filter(_filter, batched=True)
            frac = small_frac
            print("Hello 2", datetime.now().strftime("%H:%M"))

        if not (0.0 < frac <= 1.0):
            raise ValueError(f"Invalid fraction {frac}. Must be in (0, 1].")

    print("Hello 3", datetime.now().strftime("%H:%M"))
    # Subsample BEFORE the 80/20 split so both train/test reflect the chosen size
    ds = ds.shuffle(seed=seed)
    n = max(1, int(len(ds) * frac))
    ds = ds.select(range(n))

    class_counts = count_labels_as_list(ds, num_classes=10)
    print(f"Class counts...: {class_counts}")

    # 80/20 split per client (still applies in single-partition mode)
    partition_train_test = ds.train_test_split(test_size=0.2, seed=seed)

    # Torch transforms
    pytorch_transforms = Compose(
        [ToTensor(), Normalize((0.5, 0.5, 0.5), (0.5, 0.5, 0.5))]
    )

    print("Hello 4", datetime.now().strftime("%H:%M"))
    def apply_transforms(batch):
        batch["img"] = [pytorch_transforms(img) for img in batch["img"]]
        return batch

    partition_train_test = partition_train_test.with_transform(apply_transforms)

    # Dataloaders
    trainloader = DataLoader(partition_train_test["train"], batch_size=batch_size, shuffle=True)
    testloader = DataLoader(partition_train_test["test"], batch_size=batch_size)
    print("Hello 5", datetime.now().strftime("%H:%M"))
    return trainloader, testloader





# IID setup
def load_data(partition_id: int, num_partitions: int, batch_size: int,    
    small_frac: float = 1.0,   # fraction for small clients
    big_frac: float = 0.05,     # fraction for big clients (use entire partition by default)
    seed: int = 42):
   
    # Only initialize `FederatedDataset` once
    global fds
    if fds is None:
        partitioner = IidPartitioner(num_partitions=num_partitions)
        fds = FederatedDataset(
            dataset="uoft-cs/cifar10",
            partitioners={"train": partitioner},
        )

    # Load this client's IID partition (HuggingFace Dataset)
    ds = fds.load_partition(partition_id)

    # Decide how much of the partition this client gets
    frac = big_frac if partition_id in BIG_CLIENT_IDS else small_frac
    if not (0.0 < frac <= 1.0):
        raise ValueError(f"Invalid fraction {frac}. Must be in (0, 1].")

    # Subsample BEFORE the 80/20 split so both train/test reflect the chosen size
    ds = ds.shuffle(seed=seed)
    n = max(1, int(len(ds) * frac))
    ds = ds.select(range(n))

    print(n)

    # 80/20 split per client
    partition_train_test = ds.train_test_split(test_size=0.2, seed=seed)

    pytorch_transforms = Compose(
        [ToTensor(), Normalize((0.5, 0.5, 0.5), (0.5, 0.5, 0.5))]
    )

    def apply_transforms(batch):
        batch["img"] = [pytorch_transforms(img) for img in batch["img"]]
        return batch

    partition_train_test = partition_train_test.with_transform(apply_transforms)

    # Dataloaders (keep your batch_size, e.g., 32)
    trainloader = DataLoader(partition_train_test["train"], batch_size=batch_size, shuffle=True)
    testloader = DataLoader(partition_train_test["test"], batch_size=batch_size)

    return trainloader, testloader



# assumes: fds, BIG_CLIENT_IDS, FederatedDataset, IidPartitioner already exist

def _group_by_label(ds, label_col="label"):
    g = defaultdict(list)
    for i, y in enumerate(ds[label_col]):
        g[y].append(i)
    return g

def equal_per_class_subsample(ds, frac: float, seed: int = 42, label_col: str = "label"):
    
    #Pick the same number per class; total = int(len(ds)*frac) exactly.
    
    rng = np.random.default_rng(seed)
    groups = _group_by_label(ds, label_col)
    num_classes = len(groups)
    n_total = max(1, int(len(ds) * frac))
    k = n_total // num_classes
    r = n_total % num_classes

    # shuffle within each class
    for idxs in groups.values():
        rng.shuffle(idxs)

    # base equal take
    selected = []
    deficits = 0
    for c, idxs in groups.items():
        take = min(k, len(idxs))
        selected.extend(idxs[:take])
        if take < k:
            deficits += (k - take)

    # distribute remainder (and any deficits) one-per-class, then fill from pooled leftovers
    classes = list(groups.keys())
    rng.shuffle(classes)

    # first pass: try to give +1 to r classes
    extra_needed = r
    for c in classes:
        if extra_needed == 0:
            break
        idxs = groups[c]
        already = min(k, len(idxs))
        if already < len(idxs):
            selected.append(idxs[already])
            extra_needed -= 1

    # pool leftovers to make up any remaining (r not fully placed) + deficits
    need = extra_needed + deficits
    if need > 0:
        pool = []
        for c, idxs in groups.items():
            base = min(k, len(idxs))
            start = base + (1 if c in classes[:r] and base < len(idxs) else 0)
            pool.extend(idxs[start:])
        rng.shuffle(pool)
        selected.extend(pool[:need])

    # final guard: trim or top-up (top-up should never happen with CIFAR-10 IID)
    if len(selected) > n_total:
        rng.shuffle(selected)
        selected = selected[:n_total]
    elif len(selected) < n_total:
        # extremely unlikely; fill from any remaining not chosen
        all_idxs = set(range(len(ds)))
        remaining = list(all_idxs - set(selected))
        rng.shuffle(remaining)
        selected.extend(remaining[: n_total - len(selected)])

    return ds.select(selected)

def equal_per_class_split(ds, test_size: float = 0.2, seed: int = 42, label_col: str = "label"):
    
    #Per-class split with equal-per-class counts and exact overall sizes.
    
    rng = np.random.default_rng(seed)
    n = len(ds)
    n_test = int(round(n * test_size))
    groups = _group_by_label(ds, label_col)
    num_classes = len(groups)

    # shuffle per class
    for idxs in groups.values():
        rng.shuffle(idxs)

    # equal-per-class base + largest remainder to hit exact n_test
    k = len(next(iter(groups.values())))  # each class should have same size after subsample
    # If not exactly equal (just in case), fall back to per-class k_i
    per_class_sizes = {c: len(idxs) for c, idxs in groups.items()}
    equal = len(set(per_class_sizes.values())) == 1
    if equal:
        k = per_class_sizes[next(iter(per_class_sizes))]

        base = int(np.floor(k * test_size))
        remainders = {c: (k * test_size - base) for c in groups}
        test_alloc = {c: base for c in groups}
        remaining = n_test - sum(test_alloc.values())
        for c in sorted(remainders, key=remainders.get, reverse=True)[:remaining]:
            test_alloc[c] += 1
    else:
        # fallback: proportional largest remainder (rare)
        total = sum(per_class_sizes.values())
        test_alloc = {c: int(np.floor(per_class_sizes[c] * n_test / total)) for c in groups}
        remainders = {c: (per_class_sizes[c] * n_test / total) - test_alloc[c] for c in groups}
        remaining = n_test - sum(test_alloc.values())
        for c in sorted(remainders, key=remainders.get, reverse=True)[:remaining]:
            test_alloc[c] += 1

    test_idx, train_idx = [], []
    for c, idxs in groups.items():
        t = min(test_alloc[c], len(idxs))
        test_idx.extend(idxs[:t])
        train_idx.extend(idxs[t:])

    # exact guards
    if len(test_idx) < n_test:
        short = n_test - len(test_idx)
        rng.shuffle(train_idx)
        test_idx.extend(train_idx[:short])
        train_idx = train_idx[short:]
    elif len(test_idx) > n_test:
        extra = len(test_idx) - n_test
        rng.shuffle(test_idx)
        give_back = test_idx[:extra]
        train_idx.extend(give_back)
        test_idx = test_idx[extra:]

    return {"train": ds.select(train_idx), "test": ds.select(test_idx)}

def load_data(partition_id: int, num_partitions: int, batch_size: int,
              small_frac: float = 0.02, big_frac: float = 1.0, seed: int = 42):
    print("partition id, ", partition_id)

    global fds
    if fds is None:
        partitioner = IidPartitioner(num_partitions=num_partitions)
        fds = FederatedDataset(
            dataset="uoft-cs/cifar10",
            partitioners={"train": partitioner},
        )

    ds = fds.load_partition(partition_id)

    frac = big_frac if partition_id in BIG_CLIENT_IDS else small_frac
    if not (0.0 < frac <= 1.0):
        raise ValueError(f"Invalid fraction {frac}. Must be in (0, 1].")

    # Equal-per-class subset, exact size = int(len(ds) * frac)
    ds = equal_per_class_subsample(ds, frac=frac, seed=seed, label_col="label")

    # Equal-per-class 80/20 split with exact totals
    partition_train_test = equal_per_class_split(ds, test_size=0.2, seed=seed, label_col="label")

    # transforms + loaders
    pytorch_transforms = Compose([ToTensor(), Normalize((0.5, 0.5, 0.5), (0.5, 0.5, 0.5))])

    def apply_transforms(batch):
        batch["img"] = [pytorch_transforms(img) for img in batch["img"]]
        return batch

    trainset = partition_train_test["train"].with_transform(apply_transforms)
    testset = partition_train_test["test"].with_transform(apply_transforms)
    batch_size = 32

    print(batch_size)

    trainloader = DataLoader(trainset, batch_size=batch_size, shuffle=True)
    testloader = DataLoader(testset, batch_size=batch_size)

    return trainloader, testloader
"""

def train(net, trainloader, valloader, epochs, learning_rate, device):
    """Train the model on the training set."""
    global class_counts
    net.to(device)
    learning_rate = 0.01
    criterion = torch.nn.CrossEntropyLoss()
    optimizer = torch.optim.SGD(net.parameters(), lr=learning_rate, momentum=0.9, weight_decay=5e-4)
    scheduler = torch.optim.lr_scheduler.StepLR(optimizer, step_size=20, gamma=0.1)  # Optional: reduce LR every 20 epochs
    print(learning_rate)

    # Store the model weights before training
    before_training_weights = deepcopy(get_weights(net))


    for epoch in range(epochs):
        net.train()
        running_loss = 0.0

        for batch in trainloader:
            images = batch["img"].to(device)
            labels = batch["label"].to(device)

            optimizer.zero_grad()
            outputs = net(images)
            loss = criterion(outputs, labels)
            loss.backward()
            optimizer.step()

            running_loss += loss.item()

        scheduler.step()

        print(f"Epoch {epoch+1}/{epochs}, Loss: {running_loss/len(trainloader):.4f}")

    val_loss, val_acc = test(net, valloader, device)

    results = {
        "val_loss": val_loss,
        "val_accuracy": val_acc,
    }



    # Get the model weights after training
    after_training_weights = deepcopy(get_weights(net)) 

    # Calculate the L2-norm difference between the weights before and after training
    model_difference = np.sqrt(
            sum(np.sum((w_after - w_before) ** 2)
                for w_after, w_before in zip(after_training_weights, before_training_weights))
        )

    print(f"Model difference (L2-norm) after training: {model_difference}")
    print(f"Class counts...: {class_counts}")

    return results

"""
def test(net, testloader, device):
   
    net.to(device)
    net.eval()
    criterion = torch.nn.CrossEntropyLoss()
    correct, total, total_loss = 0, 0, 0.0

    with torch.no_grad():
        for batch in testloader:
            images = batch["img"].to(device)
            labels = batch["label"].to(device)

            outputs = net(images)
            loss = criterion(outputs, labels)

            total_loss += loss.item()
            _, predicted = outputs.max(1)
            correct += predicted.eq(labels).sum().item()
            total += labels.size(0)

    accuracy = correct / total
    avg_loss = total_loss / len(testloader)
    return avg_loss, accuracy
"""


def test(net, testloader, device):
    net.to(device).eval()
    criterion = torch.nn.CrossEntropyLoss()

    correct = 0
    total = 0
    total_loss = 0.0

    n_classes = None
    correct_per_class = None
    total_per_class = None

    with torch.no_grad():
        for batch in testloader:
            images = batch["img"].to(device)
            labels = batch["label"].to(device)

            outputs = net(images)
            loss = criterion(outputs, labels)
            total_loss += loss.item()

            _, predicted = outputs.max(1)
            correct += predicted.eq(labels).sum().item()
            total += labels.size(0)

            if n_classes is None:
                n_classes = outputs.size(1)
                correct_per_class = torch.zeros(n_classes, dtype=torch.long, device=labels.device)
                total_per_class   = torch.zeros(n_classes, dtype=torch.long, device=labels.device)

            # totals per true class (on-device)
            total_per_class += torch.bincount(labels, minlength=n_classes)

            # correct per class (on-device)
            mask = predicted.eq(labels)
            if mask.any():
                correct_per_class += torch.bincount(labels[mask], minlength=n_classes)

    accuracy = correct / total
    avg_loss = total_loss / len(testloader)

    # Print per-class accuracy
    class_names = getattr(getattr(testloader, "dataset", None), "classes",
                          [str(i) for i in range(n_classes)])
    per_class_acc = (correct_per_class.float() /
                     total_per_class.clamp_min(1).float())

    print("Per-class accuracy:")
    for i, name in enumerate(class_names):
        cnt = int(total_per_class[i].item())
        if cnt == 0:
            print(f"{name}: n=0 (no samples)")
        else:
            print(f"{name}: {per_class_acc[i].item()*100:.2f}%  (n={cnt})")

    return avg_loss, accuracy


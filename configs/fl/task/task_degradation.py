# lr = 0.01, batch_size = 32, epochs = 3
from collections import OrderedDict
import torch
import torch.nn as nn
import torch.nn.functional as F
from flwr_datasets import FederatedDataset
from flwr_datasets.partitioner import IidPartitioner, PathologicalPartitioner
from torch.utils.data import DataLoader
from torchvision.transforms import Compose, Normalize, ToTensor


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


fds_central = None       # for full-dataset (non-federated) case
fds_fed_full = None      # for special federated client (all labels, IID)
fds_fed_noniid = None    # for other federated clients (5 labels, non-IID)

# Common transforms
_transforms = Compose([
    ToTensor(),
    Normalize((0.5, 0.5, 0.5), (0.5, 0.5, 0.5)),
])

def _apply_transforms(batch):
    batch["img"] = [_transforms(img) for img in batch["img"]]
    return batch


def load_data_full_dataset(batch_size: int):
    """Use the entire CIFAR-10 train split (80/20 train/test)."""
    global fds_central

    if fds_central is None:
        # No partitioners -> we get the raw splits
        fds_central = FederatedDataset(
            dataset="uoft-cs/cifar10",
        )

    full_train = fds_central.load_split("train")
    split = full_train.train_test_split(test_size=0.2, seed=42)

    train_ds = split["train"].with_transform(_apply_transforms)
    test_ds = split["test"].with_transform(_apply_transforms)

    trainloader = DataLoader(train_ds, batch_size=batch_size, shuffle=True)
    testloader = DataLoader(test_ds, batch_size=batch_size)

    return trainloader, testloader

def load_data(partition_id: int, num_partitions: int, batch_size: int):

    # Global Aggregator
    if num_partitions == 0:
        return load_data_full_dataset(batch_size)

    global fds_fed_full, fds_fed_noniid

    new_client_id = 4

    # Downsampling limits for _____ clients
    max_train_samples = 200
    max_test_samples = 50

    # ----- Select FederatedDataset depending on client type -----
    if partition_id != new_client_id:
        # IID clients
        if fds_fed_full is None:
            iid_partitioner = IidPartitioner(num_partitions=num_partitions)
            fds_fed_full = FederatedDataset(
                dataset="uoft-cs/cifar10",
                partitioners={"train": iid_partitioner},
            )
        fds = fds_fed_full
    else:
        # non-IID clients
        if fds_fed_noniid is None:
            patho_partitioner = PathologicalPartitioner(
                num_partitions=num_partitions,
                partition_by="label",
                num_classes_per_partition=2,
                class_assignment_mode="random",
            )
            fds_fed_noniid = FederatedDataset(
                dataset="uoft-cs/cifar10",
                partitioners={"train": patho_partitioner},
            )
        fds = fds_fed_noniid

    # Load this client's partition
    partition = fds.load_partition(partition_id)

    # Local 80/20 split
    split = partition.train_test_split(test_size=0.2, seed=42)
    train_ds, test_ds = split["train"], split["test"]

    # Downsample only for _____ clients
    if partition_id != new_client_id:
        if max_train_samples < len(train_ds):
            train_ds = train_ds.shuffle(seed=42).select(range(max_train_samples))
        if max_test_samples < len(test_ds):
            test_ds = test_ds.shuffle(seed=42).select(range(max_test_samples))

    # Apply transforms and build dataloaders
    train_ds = train_ds.with_transform(_apply_transforms)
    test_ds = test_ds.with_transform(_apply_transforms)

    trainloader = DataLoader(train_ds, batch_size=batch_size, shuffle=True)
    testloader = DataLoader(test_ds, batch_size=batch_size)

    print("Num of batches in train set: ", len(trainloader))

    return trainloader, testloader



def train(net, trainloader, valloader, epochs, learning_rate, device):
    """Train the model on the training set."""
    net.to(device)
    criterion = torch.nn.CrossEntropyLoss()
    optimizer = torch.optim.SGD(net.parameters(), lr=learning_rate, momentum=0.9, weight_decay=5e-4)
    scheduler = torch.optim.lr_scheduler.StepLR(optimizer, step_size=20, gamma=0.1)  # Optional: reduce LR every 20 epochs

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
    return results


def test(net, testloader, device):
    """Validate the model on the test set."""
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

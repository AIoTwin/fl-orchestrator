#
# MNIST - IID
# lr = 0.01, local-epochs = 1, batch-size = 32
from collections import OrderedDict
import torch
import torch.nn as nn
import torch.nn.functional as F
from flwr_datasets import FederatedDataset
from flwr_datasets.partitioner import IidPartitioner
from torch.utils.data import DataLoader
from torchvision.transforms import Compose, Normalize, ToTensor
from torchvision import datasets, transforms
from torch.utils.data import DataLoader, Subset, random_split
from torch.utils.data import Dataset

class DictDataset(Dataset):
    """Wrap a PyTorch dataset to return dicts instead of tuples."""
    def __init__(self, base_dataset):
        self.base = base_dataset

    def __len__(self):
        return len(self.base)

    def __getitem__(self, idx):
        image, label = self.base[idx]
        return {"img": image, "label": label}

class Net(nn.Module):
    """Simple CNN for MNIST (1 input channel, 10 output classes)"""

    def __init__(self):
        super(Net, self).__init__()
        self.conv1 = nn.Conv2d(1, 6, 5)
        self.pool = nn.MaxPool2d(2, 2)
        self.conv2 = nn.Conv2d(6, 16, 5)
        self.fc1 = nn.Linear(16 * 4 * 4, 120)  # (28−5+1)//2 = 12; (12−5+1)//2 = 4
        self.fc2 = nn.Linear(120, 84)
        self.fc3 = nn.Linear(84, 10)

    def forward(self, x):
        x = self.pool(F.relu(self.conv1(x)))
        x = self.pool(F.relu(self.conv2(x)))
        x = x.view(-1, 16 * 4 * 4)
        x = F.relu(self.fc1(x))
        x = F.relu(self.fc2(x))
        return self.fc3(x)


def get_weights(net):
    return [val.cpu().numpy() for _, val in net.state_dict().items()]


def set_weights(net, parameters):
    params_dict = zip(net.state_dict().keys(), parameters)
    state_dict = OrderedDict({k: torch.tensor(v) for k, v in params_dict})
    net.load_state_dict(state_dict, strict=True)


full_dataset = None  # Cache dataset like `fds`

def load_data(partition_id: int, num_partitions: int, batch_size: int):
    """Load partitioned MNIST data manually, IID."""
    global full_dataset
    if full_dataset is None:
        transform = transforms.Compose([
            transforms.ToTensor(),
            transforms.Normalize((0.1307,), (0.3081,))
        ])
        full_dataset = datasets.MNIST(
            root="./data",
            train=True,
            download=True,
            transform=transform
        )

    # Manual IID partition
    total_size = len(full_dataset)
    partition_size = total_size // num_partitions
    start = partition_id * partition_size
    end = start + partition_size
    subset_indices = list(range(start, end))
    subset = Subset(full_dataset, subset_indices)

    # 80/20 train/test split
    train_size = int(0.8 * len(subset))
    val_size = len(subset) - train_size
    train_subset, val_subset = random_split(subset, [train_size, val_size], generator=torch.Generator().manual_seed(42))

    trainloader = DataLoader(DictDataset(train_subset), batch_size=batch_size, shuffle=True)
    testloader = DataLoader(DictDataset(val_subset), batch_size=batch_size)

    return trainloader, testloader


def train(net, trainloader, valloader, epochs, learning_rate, device):
    """Train the model on the training set."""
    net.to(device)
    criterion = torch.nn.CrossEntropyLoss().to(device)
    optimizer = torch.optim.SGD(net.parameters(), lr=learning_rate, momentum=0.9)
    net.train()
    for _ in range(epochs):
        for batch in trainloader:
            images = batch["img"]
            labels = batch["label"]
            optimizer.zero_grad()
            outputs = net(images.to(device))
            loss = criterion(outputs, labels.to(device))
            loss.backward()
            optimizer.step()

    val_loss, val_acc = test(net, valloader, device)

    results = {
        "val_loss": val_loss,
        "val_accuracy": val_acc,
    }
    return results


def test(net, testloader, device):
    """Evaluate the model on the test set."""
    net.to(device)
    criterion = torch.nn.CrossEntropyLoss()
    correct, loss = 0, 0.0
    with torch.no_grad():
        for batch in testloader:
            images = batch["img"].to(device)
            labels = batch["label"].to(device)
            outputs = net(images)
            loss += criterion(outputs, labels).item()
            correct += (outputs.argmax(1) == labels).sum().item()
    accuracy = correct / len(testloader.dataset)
    loss = loss / len(testloader)
    return loss, accuracy
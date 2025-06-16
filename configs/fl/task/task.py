#
# Cifar10 - IID
# lr = 0.1, batch_size = 32, epochs = 3
#
from collections import OrderedDict
import torch
import torch.nn as nn
import torch.nn.functional as F
from flwr_datasets import FederatedDataset
from flwr_datasets.partitioner import IidPartitioner
from torch.utils.data import DataLoader
from torchvision.transforms import Compose, Normalize, ToTensor

from torch.utils.data import DataLoader, Subset, random_split
from torchvision.datasets import CIFAR10
from torchvision.transforms import Compose, ToTensor, Normalize


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

def load_data(partition_id: int, num_partitions: int, batch_size: int):
    """Load partitioned CIFAR-10 data using torchvision with manual partitioning."""
    
    transform = Compose([
        ToTensor(),
        Normalize((0.5, 0.5, 0.5), (0.5, 0.5, 0.5))
    ])

    # Load the full CIFAR10 training set
    full_dataset = CIFAR10(root="./data", train=True, download=True, transform=transform)
    total_size = len(full_dataset)

    # Partition the dataset (iid assumption)
    partition_size = total_size // num_partitions
    start_idx = partition_id * partition_size
    end_idx = start_idx + partition_size
    indices = list(range(start_idx, end_idx))

    # Get subset for the current partition
    partition_dataset = Subset(full_dataset, indices)

    # Split into train/test (e.g., 80% train, 20% test)
    train_size = int(0.8 * len(partition_dataset))
    test_size = len(partition_dataset) - train_size
    train_subset, test_subset = random_split(partition_dataset, [train_size, test_size], generator=torch.Generator().manual_seed(42))

    # Create DataLoaders
    trainloader = DataLoader(train_subset, batch_size=batch_size, shuffle=True)
    testloader = DataLoader(test_subset, batch_size=batch_size, shuffle=False)

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
            images, labels = batch
            images = images.to(device)
            labels = labels.to(device)


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
            images, labels = batch
            images = images.to(device)
            labels = labels.to(device)

            outputs = net(images)
            loss = criterion(outputs, labels)

            total_loss += loss.item()
            _, predicted = outputs.max(1)
            correct += predicted.eq(labels).sum().item()
            total += labels.size(0)

    accuracy = correct / total
    avg_loss = total_loss / len(testloader)
    return avg_loss, accuracy


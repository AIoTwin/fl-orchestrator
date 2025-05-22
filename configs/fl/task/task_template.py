from collections import OrderedDict

import torch
import torch.nn as nn
from torch.utils.data import DataLoader
from typing import Tuple, List, Dict, Any


class Net(nn.Module):
    def __init__(self) -> None:
        ...

    def forward(self, x: torch.Tensor) -> torch.Tensor:
        ...


def get_weights(net: nn.Module) -> List[Any]:
    ...


def set_weights(net: nn.Module, parameters: List[Any]) -> None:
    ...


def load_data(partition_id: int, num_partitions: int, batch_size: int) -> Tuple[DataLoader, DataLoader]:
    ...


def train(
    net: nn.Module,
    trainloader: DataLoader,
    valloader: DataLoader,
    epochs: int,
    learning_rate: float,
    device: torch.device
) -> Dict[str, float]:
    ...


def test(net: nn.Module, testloader: DataLoader, device: torch.device) -> Tuple[float, float]:
    ...

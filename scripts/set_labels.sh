# k3s-master-1
kubectl label --overwrite nodes k3s-master-1 fl/type=global_aggregator

# k3s-node-1
kubectl label --overwrite nodes k3s-node-1 fl/type=local_aggregator
kubectl label --overwrite nodes k3s-node-1 comm/k3s-master-1=100

# k3s-node-2
kubectl label --overwrite nodes k3s-node-2 fl/type=local_aggregator
kubectl label --overwrite nodes k3s-node-2 comm/k3s-master-1=100

# k3s-node-3
kubectl label --overwrite nodes k3s-node-3 fl/type=client
kubectl label --overwrite nodes k3s-node-3 comm/k3s-node-1=100
kubectl label --overwrite nodes k3s-node-3 comm/k3s-node-2=50
kubectl label --overwrite nodes k3s-node-3 data/0=1000
kubectl label --overwrite nodes k3s-node-3 data/1=10000
kubectl label --overwrite nodes k3s-node-3 data/2=20000

# k3s-node-4
kubectl label --overwrite nodes k3s-node-4 fl/type=client
kubectl label --overwrite nodes k3s-node-4 comm/k3s-node-1=100
kubectl label --overwrite nodes k3s-node-4 comm/k3s-node-2=100
kubectl label --overwrite nodes k3s-node-4 data/3=1000
kubectl label --overwrite nodes k3s-node-4 data/4=10000
kubectl label --overwrite nodes k3s-node-4 data/5=20000

# k3s-node-5
kubectl label --overwrite nodes k3s-node-5 fl/type=client
kubectl label --overwrite nodes k3s-node-5 comm/k3s-node-1=100
kubectl label --overwrite nodes k3s-node-5 comm/k3s-node-2=100
kubectl label --overwrite nodes k3s-node-5 data/6=1000
kubectl label --overwrite nodes k3s-node-5 data/7=10000
kubectl label --overwrite nodes k3s-node-5 data/8=20000

# k3s-node-6
kubectl label --overwrite nodes k3s-node-6 fl/type=client
kubectl label --overwrite nodes k3s-node-6 comm/k3s-node-1=50
kubectl label --overwrite nodes k3s-node-6 comm/k3s-node-2=100
kubectl label --overwrite nodes k3s-node-6 data/0=1000
kubectl label --overwrite nodes k3s-node-6 data/4=10000
kubectl label --overwrite nodes k3s-node-6 data/9=20000

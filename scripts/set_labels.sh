# k3s master
kubectl label --overwrite nodes k3s-master fl/type=global_aggregator

# survey-orch1
kubectl label --overwrite nodes survey-orch1 fl/type=local_aggregator
kubectl label --overwrite nodes survey-orch1 comm/k3s-master=100

# fer-iot
kubectl label --overwrite nodes fer-iot fl/type=local_aggregator
kubectl label --overwrite nodes fer-iot comm/k3s-master=100

# k3s-node-1
kubectl label --overwrite nodes k3s-node-1 fl/type=client
kubectl label --overwrite nodes k3s-node-1 comm/survey-orch1=100
kubectl label --overwrite nodes k3s-node-1 comm/fer-iot=50
kubectl label --overwrite nodes k3s-node-1 data/0=1000
kubectl label --overwrite nodes k3s-node-1 data/1=10000
kubectl label --overwrite nodes k3s-node-1 data/2=20000

# k3s-node-2
kubectl label --overwrite nodes k3s-node-2 fl/type=client
kubectl label --overwrite nodes k3s-node-2 comm/survey-orch1=100
kubectl label --overwrite nodes k3s-node-2 comm/fer-iot=100
kubectl label --overwrite nodes k3s-node-2 data/3=1000
kubectl label --overwrite nodes k3s-node-2 data/4=10000
kubectl label --overwrite nodes k3s-node-2 data/5=20000

# k3s-node-3
kubectl label --overwrite nodes k3s-node-3 fl/type=client
kubectl label --overwrite nodes k3s-node-3 comm/survey-orch1=100
kubectl label --overwrite nodes k3s-node-3 comm/fer-iot=100
kubectl label --overwrite nodes k3s-node-3 data/6=1000
kubectl label --overwrite nodes k3s-node-3 data/7=10000
kubectl label --overwrite nodes k3s-node-3 data/8=20000

# k3s-node-4
kubectl label --overwrite nodes k3s-node-4 fl/type=client
kubectl label --overwrite nodes k3s-node-4 comm/survey-orch1=50
kubectl label --overwrite nodes k3s-node-4 comm/fer-iot=100
kubectl label --overwrite nodes k3s-node-4 data/0=1000
kubectl label --overwrite nodes k3s-node-4 data/4=10000
kubectl label --overwrite nodes k3s-node-4 data/9=20000

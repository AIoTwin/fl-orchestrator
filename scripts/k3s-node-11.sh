# k3s-node-11
kubectl label --overwrite nodes k3s-node-11 fl/type=client
kubectl label --overwrite nodes k3s-node-11 comm/k3s-node-1=60
kubectl label --overwrite nodes k3s-node-11 comm/k3s-node-2=20
kubectl label --overwrite nodes k3s-node-11 data/6=1000
kubectl label --overwrite nodes k3s-node-11 data/7=1000
kubectl label --overwrite nodes k3s-node-11 data/8=1000

echo $(date +"%Y-%m-%d %H:%M:%S.%3N")
kubectl label nodes k3s-master fl/type=aggregator
kubectl label nodes survey-orch1 fl/type=client
kubectl label nodes fer-iot fl/type=client

kubectl label nodes survey-orch1 comm/k3s-master=100
kubectl label nodes survey-orch1 data/0=100
kubectl label nodes survey-orch1 data/1=1000

kubectl label nodes fer-iot comm/k3s-master=100
kubectl label nodes fer-iot data/2=10000
kubectl label nodes fer-iot data/3=20000

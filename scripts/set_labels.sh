# global aggregator
kubectl label --overwrite nodes ga fl/type=global_aggregator

# local aggregator
kubectl label --overwrite nodes la-1 fl/type=local_aggregator
kubectl label --overwrite nodes la-1 comm/ga=100

# client
kubectl label --overwrite nodes cl-1 fl/type=client
kubectl label --overwrite nodes cl-1 comm/ga=120
kubectl label --overwrite nodes cl-1 comm/la-1=20
kubectl label --overwrite nodes cl-1 comm/la-2=60
kubectl label --overwrite nodes cl-1 fl/num-partitions=10
kubectl label --overwrite nodes cl-1 fl/partition-id=0

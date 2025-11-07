# global aggregator
kubectl label --overwrite nodes hfl-n1 fl/type=global_aggregator

# local aggregator
kubectl label --overwrite nodes hfl-n2 fl/type=local_aggregator
kubectl label --overwrite nodes hfl-n2 comm/hfl-n1=10

kubectl label --overwrite nodes mon0 fl/type=local_aggregator
kubectl label --overwrite nodes mon0 comm/hfl-n1=100

# client
kubectl label --overwrite nodes hfl-n3 fl/type=client
kubectl label --overwrite nodes hfl-n3 comm/hfl-n2=10
kubectl label --overwrite nodes hfl-n3 comm/mon0=100
kubectl label --overwrite nodes hfl-n3 fl/num-partitions=8
kubectl label --overwrite nodes hfl-n3 fl/partition-id=0

kubectl label --overwrite nodes hfl-n4 fl/type=client
kubectl label --overwrite nodes hfl-n4 comm/hfl-n2=10
kubectl label --overwrite nodes hfl-n4 comm/mon0=100
kubectl label --overwrite nodes hfl-n4 fl/num-partitions=8
kubectl label --overwrite nodes hfl-n4 fl/partition-id=1

kubectl label --overwrite nodes hfl-n5 fl/type=client
kubectl label --overwrite nodes hfl-n5 comm/hfl-n2=10
kubectl label --overwrite nodes hfl-n5 comm/mon0=100
kubectl label --overwrite nodes hfl-n5 fl/num-partitions=8
kubectl label --overwrite nodes hfl-n5 fl/partition-id=2

kubectl label --overwrite nodes hfl-n6 fl/type=client
kubectl label --overwrite nodes hfl-n6 comm/hfl-n2=10
kubectl label --overwrite nodes hfl-n6 comm/mon0=100
kubectl label --overwrite nodes hfl-n6 fl/num-partitions=8
kubectl label --overwrite nodes hfl-n6 fl/partition-id=3

kubectl label --overwrite nodes mon1 fl/type=client
kubectl label --overwrite nodes mon1 comm/hfl-n2=100
kubectl label --overwrite nodes mon1 comm/mon0=10
kubectl label --overwrite nodes mon1 fl/num-partitions=8
kubectl label --overwrite nodes mon1 fl/partition-id=4

kubectl label --overwrite nodes mon2 fl/type=client
kubectl label --overwrite nodes mon2 comm/hfl-n2=100
kubectl label --overwrite nodes mon2 comm/mon0=10
kubectl label --overwrite nodes mon2 fl/num-partitions=8
kubectl label --overwrite nodes mon2 fl/partition-id=5

kubectl label --overwrite nodes mon3 fl/type=client
kubectl label --overwrite nodes mon3 comm/hfl-n2=100
kubectl label --overwrite nodes mon3 comm/mon0=10
kubectl label --overwrite nodes mon3 fl/num-partitions=8
kubectl label --overwrite nodes mon3 fl/partition-id=6

kubectl label --overwrite nodes mon4 fl/type=client
kubectl label --overwrite nodes mon4 comm/hfl-n2=100
kubectl label --overwrite nodes mon4 comm/mon0=10
kubectl label --overwrite nodes mon4 fl/num-partitions=8
kubectl label --overwrite nodes mon4 fl/partition-id=7
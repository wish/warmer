# warmer [![Docker Repository on Quay](https://quay.io/repository/wish/warmer/status "Docker Repository on Quay")](https://quay.io/repository/wish/warmer)

warmer is utility program that warms EBS drive that was restored from snapshot
by reading all files in order of their physical location on EBS to maximize
throughput on restore.

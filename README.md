# nrm
NATS Streaming: Extract random (or defined) message from subject log

## Install

`go get github.com/larskluge/nrm`

## Usage

### Print Usage

```
$ nrm
Specify filter; format: <subject>:<sequence number>, only subject is required
```

### Random Message

```
$ nrm foo
Seq 1-20056; Picked: foo:7266
{"bar":"baz"}
```

### Specified Message

```
$ nrm foo:7266
Picked: foo:7266
{"bar":"baz"}
```

### Chain w/ other Commands

```
nrm foo | jq .
```

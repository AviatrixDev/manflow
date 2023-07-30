# manflow

Netflow generator

Documentation: https://aviatrix.atlassian.net/wiki/spaces/copconf/pages/1913815177/manflow+netflow+generator

## Usage

Using following `flowConfig.json` file:

```json
{
  "seed": 1,
  "flow_timeout": 10,
  "collector_ip": "10.0.0.11",
  "collector_port": 31283,
  "hosts": [
    {
      "ip": "10.0.0.103",
      "name": "gw1"
    }
  ],
  "flows": [
    {
      "src_addr": "10.14.0.0",
      "dst_addr": "10.0.1.0",
      "dst_port": "80",
      "hops": ["gw1"],
      "count": 1
    }
  ]
}
```

Run following commands

```bash
go build # one time
./manflow -i gw1
```

Command-line arguments:

- `-i` - host name of one of the hosts in `flowConfig.json` file
- `-l` - disable flow-level logging

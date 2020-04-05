# opennet
cni plugin for assign POD IP , cluster level, can fix POD IP



Intro

```json
{
    "dev":"ip address",
    "cni":"cni configuration",
    "containerID":"containerID"
}
```



IPAM

```json
{
  "cniVersion": "0.3.1",
  "name": "opennet-net",
  "type": "macvlan",
  "master": "eno1",
  "mode": "vepa",
  "ipam": {
    "type": "opennet",
    "range": [
        {
          "type":"all",
          "vals": [
                {
                    "start":"23.106.248.1",
                    "end":"23.106.248.250",
                    "gw	":"23.106.248.190"
                },
                {
                    "start":"23.106.252.1",
                    "end":"23.106.252.250",
                    "gw":"23.106.252.126"
                }
            ] 
        },
        {
           "type":"23.106.248.115",
           "vals":[
               {
                    "start":"23.106.240.1",
                    "end":"23.106.240.250",
                    "gw":"23.106.240.190"
                },
                {
                    "start":"23.106.253.1",
                    "end":"23.106.253.250",
                    "gw":"23.106.253.126"
                }
           ]
        }
    ],
    "dns": {
      "nameserver": [
        "8.8.8.8",
        "8.8.4.4"
      ]
    }
  }
}
```


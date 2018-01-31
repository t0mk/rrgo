#!/usr/bin/env python

import json

moar = [("WETH", "0x2956356cd2a2bf3202f771f50d3d14a367b48070"),
        ("NEWB", "0x814964b1bceAf24e26296D031EaDf134a2Ca4105")]


ethTokens = json.loads(open("ethTokens.json", "r").read())

d = {i['symbol']: i['address'].lower() for i in ethTokens}

d.update({s: a.lower() for s, a in moar})

print("""
package rrgo

var (
        A2T = map[string]string{
""")

for s,a in d.items():
    print(u'"{}": "{}",'.format(a, s).encode("utf-8"))
print("""
}
        T2A = map[string]string{
""")

for s,a in d.items():
    print(u'"{}": "{}",'.format(s, a).encode("utf-8"))
print("})")


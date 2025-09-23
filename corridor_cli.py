#!/usr/bin/env python3
import argparse, json, sys, urllib.request

BASE = "http://localhost:7080"

def post(path, payload=None):
    data = None if payload is None else json.dumps(payload).encode()
    req = urllib.request.Request(BASE+path, data=data, headers={"Content-Type":"application/json"})
    with urllib.request.urlopen(req) as resp:
        return json.loads(resp.read().decode())

def get(path):
    with urllib.request.urlopen(BASE+path) as resp:
        return json.loads(resp.read().decode())

def cmd_corridor_alloc(args):
    payload = {
        "type": args.type, "lanes": args.lanes,
        "lambda_nm": list(range(args.lambda_start, args.lambda_start+args.lanes)),
        "min_gbps": args.min_gbps, "latency_budget_ns": args.latency_ns,
        "reach_mm": args.reach_mm, "mode": "waveguide",
        "qos": {"pfc": True, "priority": "gold"}, "attestation_required": True
    }
    print(json.dumps(post("/v1/corridors", payload), indent=2))

def cmd_ffm_alloc(args):
    payload = {
        "bytes": args.bytes, "latency_class": args.tier,
        "bandwidth_floor_GBs": args.bw_floor, "persistence": "none",
        "shareable": True, "security_domain": "tenantA"
    }
    print(json.dumps(post("/v1/ffm/alloc", payload), indent=2))

def main():
    p = argparse.ArgumentParser(prog="corridor")
    sub = p.add_subparsers()

    c1 = sub.add_parser("lanes-alloc")
    c1.add_argument("--type", default="SiCorridor")
    c1.add_argument("--lanes", type=int, default=8)
    c1.add_argument("--lambda-start", type=int, default=1550)
    c1.add_argument("--min-gbps", type=int, default=400)
    c1.add_argument("--latency-ns", type=int, default=250)
    c1.add_argument("--reach-mm", type=int, default=75)
    c1.set_defaults(func=cmd_corridor_alloc)

    c2 = sub.add_parser("ffm-alloc")
    c2.add_argument("--bytes", type=int, default=256*1024*1024*1024)
    c2.add_argument("--tier", default="T2")
    c2.add_argument("--bw-floor", type=int, default=150)
    c2.set_defaults(func=cmd_ffm_alloc)

    args = p.parse_args()
    if hasattr(args, "func"): args.func(args)
    else: p.print_help()

if __name__ == "__main__":
    main()

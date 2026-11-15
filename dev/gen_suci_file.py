#!/usr/bin/env python3
"""Build a file for `packetrusher multi-ue --suciFile`.

Each output line is one hex-encoded 5GS mobile identity (SUCI, IMSI format).
Null-scheme MSINs are BCD-encoded; for other schemes give --scheme-output
(or the scheme_output CSV column) as raw hex.

Generate N sequential SUCIs (MSIN counts up, width preserved):
  gen_suci_file.py --mcc 999 --mnc 70 --msin 0000000120 --count 5 -o sucis.txt

Build from a CSV (header: mcc,mnc,routing_indicator,scheme,key_id,msin,scheme_output;
msin and scheme_output may be empty, scheme_output wins):
  gen_suci_file.py --csv rows.csv -o sucis.txt

Several SUCIs per SUPI (here 3, each with a different random 40-48 byte scheme
output, as ECIES would produce):
  gen_suci_file.py --scheme 1 --count 5 --repeat 3 --random-scheme-output 40-48

Random bit streams, 20 lines of 1-64 bytes, sent verbatim even if not a valid
identity (add --seed for a reproducible file):
  gen_suci_file.py --random-raw 20 --length 1-64

Show what a file contains:
  gen_suci_file.py --decode sucis.txt
"""
import argparse
import csv
import random
import sys


def bcd(digits: str) -> bytes:
    """Pack digits two per octet, first digit in the low nibble, pad with F."""
    if not digits or any(c not in "0123456789f" for c in digits):
        raise ValueError(f"not all digits: {digits!r}")
    if len(digits) % 2:
        digits += "f"
    return bytes(int(digits[i + 1], 16) << 4 | int(digits[i], 16) for i in range(0, len(digits), 2))


def plmn(mcc: str, mnc: str) -> bytes:
    if len(mcc) != 3 or not mcc.isdigit():
        raise ValueError(f"MCC must be 3 digits: {mcc!r}")
    if len(mnc) not in (2, 3) or not mnc.isdigit():
        raise ValueError(f"MNC must be 2 or 3 digits: {mnc!r}")
    mnc3 = mnc[2] if len(mnc) == 3 else "f"
    return bytes([
        int(mcc[1]) << 4 | int(mcc[0]),
        int(mnc3, 16) << 4 | int(mcc[2]),
        int(mnc[1]) << 4 | int(mnc[0]),
    ])


def routing(ri: str) -> bytes:
    if len(ri) > 4 or not ri.isdigit():
        raise ValueError(f"routing indicator must be 1-4 digits: {ri!r}")
    return bcd(ri.ljust(4, "f"))


def build(mcc, mnc, ri, scheme, key_id, msin=None, scheme_output=None) -> bytes:
    if msin is not None and msin and not msin.isdigit():
        raise ValueError(f"MSIN must be digits: {msin!r}")
    if not 0 <= scheme <= 15:
        raise ValueError(f"scheme must be 0-15: {scheme}")
    if not 0 <= key_id <= 255:
        raise ValueError(f"key id must be 0-255: {key_id}")
    if scheme_output is not None and scheme_output != "":
        out = bytes.fromhex(scheme_output.replace(" ", "").removeprefix("0x"))
    elif msin:
        out = bcd(msin)
    else:
        raise ValueError("need an msin or a scheme output")
    return bytes([0x01]) + plmn(mcc, mnc) + routing(ri) + bytes([scheme, key_id]) + out


def decode(raw: bytes) -> str:
    if len(raw) < 8 or raw[0] & 0x07 != 1:
        return "not a SUCI: " + raw.hex()
    p = raw[1:4]
    mcc = f"{p[0] & 15}{p[0] >> 4}{p[1] & 15}"
    mnc = f"{p[2] & 15}{p[2] >> 4}" + ("" if p[1] >> 4 == 15 else f"{p[1] >> 4}")
    ri = "".join(f"{b & 15:x}{b >> 4:x}" for b in raw[4:6]).rstrip("f")
    out = raw[8:]
    msin = "".join(f"{b & 15:x}{b >> 4:x}" for b in out).rstrip("f")
    shown = f"msin {msin}" if raw[6] & 15 == 0 else f"scheme output {out.hex()}"
    return f"mcc {mcc} mnc {mnc} ri {ri or '-'} scheme {raw[6] & 15} key {raw[7]} {shown}"


def span(text: str):
    lo, _, hi = text.partition("-")
    lo, hi = int(lo), int(hi or lo)
    if lo < 0 or hi < lo:
        raise ValueError(f"bad length range: {text!r}")
    return lo, hi


def rand_bytes(rng, lo, hi) -> bytes:
    return rng.randbytes(rng.randint(lo, hi))


def read_hex_lines(path):
    for line in open(path):
        line = line.split("#")[0].strip().lower().removeprefix("0x")
        line = line.replace(" ", "").replace(":", "")
        if line:
            yield bytes.fromhex(line)


def main():
    ap = argparse.ArgumentParser(description="Build a --suciFile for packetrusher multi-ue.",
                                 formatter_class=argparse.RawDescriptionHelpFormatter, epilog=__doc__)
    ap.add_argument("-o", "--output", help="output file (default: stdout)")
    ap.add_argument("--decode", metavar="FILE", help="print the fields of each line in FILE and exit")
    ap.add_argument("--csv", metavar="FILE", help="build from CSV rows instead of generating")
    ap.add_argument("--count", type=int, default=1, help="number of SUCIs to generate (default 1)")
    ap.add_argument("--mcc", default="999")
    ap.add_argument("--mnc", default="70")
    ap.add_argument("--routing-indicator", default="0000")
    ap.add_argument("--scheme", type=int, default=0, help="protection scheme id (default 0, null)")
    ap.add_argument("--key-id", type=int, default=0, help="home network public key id (default 0)")
    ap.add_argument("--msin", default="0000000120", help="first MSIN (default 0000000120)")
    ap.add_argument("--scheme-output", help="raw hex scheme output, used as-is for every line")
    ap.add_argument("--repeat", type=int, default=1, help="lines per SUPI (default 1)")
    ap.add_argument("--random-scheme-output", metavar="LEN|MIN-MAX",
                    help="random bytes as the scheme output of each line")
    ap.add_argument("--random-raw", type=int, metavar="COUNT",
                    help="emit COUNT lines of random bytes, no SUCI structure")
    ap.add_argument("--length", default="8-64", metavar="LEN|MIN-MAX",
                    help="byte length range for --random-raw (default 8-64)")
    ap.add_argument("--seed", type=int, help="seed for the random options")
    a = ap.parse_args()

    if a.decode:
        for n, raw in enumerate(read_hex_lines(a.decode), 1):
            print(f"{n}: {decode(raw)}")
        return

    rng = random.Random(a.seed)
    lines = []
    try:
        if a.random_raw is not None:
            lo, hi = span(a.length)
            lines = [rand_bytes(rng, max(lo, 1), max(hi, 1)) for _ in range(a.random_raw)]
        elif a.csv:
            with open(a.csv, newline="") as f:
                for n, row in enumerate(csv.DictReader(f), 2):
                    try:
                        lines.append(build(row["mcc"].strip(), row["mnc"].strip(), row["routing_indicator"].strip(),
                                           int(row["scheme"] or 0), int(row["key_id"] or 0),
                                           (row.get("msin") or "").strip(), (row.get("scheme_output") or "").strip()))
                    except (ValueError, KeyError) as e:
                        sys.exit(f"{a.csv}:{n}: {e}")
        else:
            width = len(a.msin)
            rso = span(a.random_scheme_output) if a.random_scheme_output else None
            for i in range(a.count):
                msin = str(int(a.msin) + i).zfill(width)
                for _ in range(a.repeat):
                    out = rand_bytes(rng, *rso).hex() if rso else a.scheme_output
                    lines.append(build(a.mcc, a.mnc, a.routing_indicator, a.scheme, a.key_id,
                                       None if rso else msin, out))
    except ValueError as e:
        sys.exit(str(e))

    text = "".join(f"{raw.hex()}  # {decode(raw)}\n" for raw in lines)
    if a.output:
        with open(a.output, "w") as f:
            f.write(text)
        print(f"wrote {len(lines)} SUCI(s) to {a.output}", file=sys.stderr)
    else:
        sys.stdout.write(text)


if __name__ == "__main__":
    main()

#!/usr/bin/env python3
"""Dado un conjunto de ficheros cambiados, imprime qué owners exige CODEOWNERS
(último patrón que casa gana, como en GitHub). Demuestra el enrutado de revisión."""
import sys, fnmatch

def load_rules(path="CODEOWNERS"):
    rules = []
    for line in open(path, encoding="utf-8"):
        line = line.strip()
        if not line or line.startswith("#"):
            continue
        pat, *owners = line.split()
        rules.append((pat, owners))
    return rules

def owners_for(path, rules):
    winner = None
    for pat, owners in rules:
        p = pat.lstrip("/")
        # patrón de directorio: casa por prefijo
        if p.endswith("/"):
            if path.startswith(p):
                winner = owners
        elif fnmatch.fnmatch(path, p) or fnmatch.fnmatch(path, p + "/*"):
            winner = owners
    return winner

rules = load_rules()
files = sys.argv[1:]
req = {}
for f in files:
    ow = owners_for(f, rules)
    if ow:
        req[tuple(ow)] = req.get(tuple(ow), 0) + 1
print("  Revisores requeridos por CODEOWNERS:")
for owners, n in req.items():
    print(f"    {' '.join(owners):28} ({n} fichero/s)")

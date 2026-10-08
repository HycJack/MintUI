import json, re, glob, sys
base=json.load(open('docs/component-catalogue.baseline.json'))
def exported(pkg):
    n=set()
    for f in glob.glob('ui/%s/*.go'%pkg):
        if f.endswith('_test.go'): continue
        for line in open(f):
            m=re.match(r'func ([A-Z]\w*)\(',line) or re.match(r'type ([A-Z]\w*)\b',line)
            if m: n.add(m.group(1))
    return n
def norm(s):
    # ChartAxis -> axis, OHLCChart -> ohlc ; VolumeChart -> volume
    s=re.sub(r'(Chart|Options|Result|State)$','',s)
    s=re.sub(r'^Chart','',s)
    return s.lower()
tot=exact=alias=0; real={}
for pkg in sorted(base):
    want=[e['name'] for e in base[pkg]]; got=exported(pkg)
    gn={norm(g) for g in got}
    ex=al=0; miss=[]
    for w in want:
        tot+=1
        if w in got: ex+=1; exact+=1
        elif norm(w) in gn: al+=1; alias+=1
        else: miss.append(w)
    real[pkg]=miss
    print('%-11s %3d/%3d  (同名 %d, 改名 %d, 缺 %d) %s'%(
        pkg,ex+al,len(want),ex,al,len(miss), ('缺 '+', '.join(miss)) if miss else '✓'))
print('\n基线 %d 项：同名 %d + 改名 %d = %d  真正缺 %d  → %.0f%%'%(
    tot,exact,alias,exact+alias,tot-exact-alias,100*(exact+alias)/tot))

#!/usr/bin/env python
 
import re

f=open("index.html")
lines = f.readlines()
f.close()

r1=re.compile('<tr>(.*?)</tr>')
r2=re.compile('<td><.*?>(.*)<.*?><.*?><.*?>(.*)</td>')

print "package main"
print ""
print "var x86ops = map[string]string{"
for l in lines:
    if "Core I" in l:
        m1=r1.findall(l) 
        for i in m1:
            m2=r2.match(i)
            if m2:
                print '"%s" : "%s",' % (m2.groups()[0].lower(), m2.groups()[1])
print "}" 

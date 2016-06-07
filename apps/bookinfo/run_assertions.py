#!/usr/bin/python
import json
import sys
from pygremlin import A8AssertionChecker

def passOrfail(result):
    if result:
        return "PASS"
    else:
        return "FAIL"

with open("checklist.json") as fp:
    checklist = json.load(fp)
ac = A8AssertionChecker(checklist['log_server'], None, header="Cookie", pattern="user=jason")
results = ac.check_assertions(checklist, all=True)
exit_status = 0

for check in results:
    print 'Check %s %s %s:%s' % (check.name, check.info, passOrfail(check.success), check.errormsg)
    if not check.success:
        exit_status = 1
sys.exit(exit_status)

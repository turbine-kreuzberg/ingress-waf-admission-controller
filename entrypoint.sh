#!/bin/sh
echo `which ingress-waf-admission-controller` | entr -nr `which ingress-waf-admission-controller` $@

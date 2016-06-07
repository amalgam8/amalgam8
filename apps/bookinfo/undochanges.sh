#!/bin/bash
./killapps.sh
cp productpage/productpage-v1.py productpage/productpage.py
./rebuild-productpage.sh

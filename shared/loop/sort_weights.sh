#!/bin/bash

sort -nk2 loop-19.txt > sorted_loop_19.txt
awk '{print $1}' sorted_loop_19.txt > loop_19_page.txt
sort -nk2 loop.txt > sorted_loop.txt
awk '{print $1}' sorted_loop.txt > loop_page.txt


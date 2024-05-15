@echo off

for /l %%i in (1,1,10) do (
    echo loop %%i
    ping 127.0.0.1 -n 2 > nul
)

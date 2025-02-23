rsrc -manifest UDT.manifest -ico icon.ico -o rsrc.syso
@REM go build -ldflags="-s -w" -o .\UDT\UDT.exe
go build -ldflags="-H windowsgui -s -w" -o .\UDT\UDT.exe
cp .\config.yaml .\UDT\config.yaml
cd manager
go build -ldflags="-s -w" -o ..\UDT\manager.exe
cd ..

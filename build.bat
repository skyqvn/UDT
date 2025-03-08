rsrc -manifest UDT.manifest -ico logo.ico -o rsrc.syso
go build -ldflags="-s -w" -o .\UDT\UDT.exe
go build -ldflags="-H windowsgui -s -w" -o .\UDT\UDT.exe
@REM cp .\config.yaml .\UDT\config.yaml
cd manager
go build -ldflags="-s -w" -o ..\UDT\manager.exe
cd ..

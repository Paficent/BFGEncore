rmdir /S /Q "..\builds"

mkdir "../builds"
mkdir "../builds/linux"
mkdir "../builds/windows"

go build -o ../builds/linux/server ./cmd/server
go build -o ../builds/linux/setup ./cmd/setup

go build -o ../builds/windows/server.exe ./cmd/server/
go build -o ../builds/windows/setup.exe ./cmd/setup/

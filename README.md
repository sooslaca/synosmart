# synosmart
Workaround for Synology DSM 7 missing S.M.A.R.T info.  

### Build
```
go build -trimpath -ldflags '-s -w'
```
Copy the binary to your Synology NAS and use it from SSH.

### Example output
![Example output](/img/example.png "Example output")

### Credits
Credit for https://github.com/anatol/smart.go for the S.M.A.R.T. golang library.
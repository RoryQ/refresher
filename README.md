# Refresher

Checks your gcloud auth token and launches the login page when expired.

### Installation

If using go 1.16+ then `go install` will place the binary on your path

```
go install github.com/roryq/refresher@v0.1.1
```

Then call the `refresher` before gcloud auth is needed. 

e.g. a kubectl alias in fish:

```fish
function k --wraps=kubectl
        refresher
        kubectl $argv
end
```

Or you could place in your .bashrc file to check when you open a terminal.
When reading from the cache it should add minimal overhead


```
> time refresher

________________________________________________________
Executed in    9.59 millis    fish           external
   usr time    2.69 millis   83.00 micros    2.61 millis
   sys time    5.13 millis  575.00 micros    4.55 millis
```

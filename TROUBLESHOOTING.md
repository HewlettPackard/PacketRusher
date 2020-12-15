## Troubleshooting

If you get this error:
```
./app: error while loading shared libraries: libgsl.so.25: cannot open shared object file: No such file or directory
```

Please check the path of the GSL dynamic library:
```
LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/usr/local/lib/gsl/
export LD_LIBRARY_PATH
./app
```

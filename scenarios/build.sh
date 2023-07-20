BASEDIR=$(dirname "$0")

for folder in $BASEDIR/*; do
  for scenario in $folder/*.go; do
    if [ -f $scenario ]; then
      tinygo build -o $BASEDIR/$(basename $scenario).wasm -target=wasi $scenario
      echo "You can now run the scenario $(basename $scenario).wasm using ./app custom-scenario --scenario $(basename $scenario).wasm"
    fi
  done
done
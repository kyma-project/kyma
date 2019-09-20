if [ ! -d ".debug" ]; then
  echo "This script expects to be run in the root directory of the project."
  exit 1
fi

bashCmd="-c .debug/runDelveLoop.sh"

if [ "$1" == "--manual" ]; then
   bashCmd=""
fi

docker build .debug -t vscode-go-debug

telepresence --docker-run -v "$(pwd)":/opt/go/src/local/myorg/myapp \
                          -p 2345:2345 -it \
                          --cap-add=SYS_PTRACE \
                          vscode-go-debug \
                          bash $bashCmd
function cacheImage () {
  regex="^(eu\.gcr\.io/|gcr\.io/|ghcr\.io/|docker\.io/|quay\.io/)(.*)"
  if [[ $1 =~ $regex ]]
  then
    registry=${BASH_REMATCH[1]//[$'\t\r\n']}
    image=${BASH_REMATCH[2]//[$'\t\r\n']}
    crane cp ${registry}${image} registry.localhost:5000/${image}
  fi
}

docker exec -it k3d-kyma-server-0 sh -c "ctr images ls -q" >images.txt 
cat images.txt | while read -r line; do cacheImage "$line"; done


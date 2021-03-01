
function queryRuntimeStatus(runtimeID) {
    return `query { result: runtimeStatus(id: \\"${runtimeID}\\") { data { runtimeConfiguration{ kubeconfig }  } } }`;
}


module.exports = {
    queryRuntimeStatus
} 
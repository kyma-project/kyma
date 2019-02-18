{
  ID: 'nodejs',
  versions:
    [ { name: 'node6',
       version: '6',
       images: [{
        phase: "installation",
        image: "kubeless/nodejs@sha256:10f6a3d3e0782220d3a20a0847130846d554e1f196898d928ed986d8a21ff00f",
        command: "/kubeless-npm-install.sh"
       }, {
        phase: "runtime",
        image: "kubeless/nodejs@sha256:10f6a3d3e0782220d3a20a0847130846d554e1f196898d928ed986d8a21ff00f",
        env: {
          NODE_PATH: "$(KUBELESS_INSTALL_VOLUME)/node_modules",
        },
       }],
      },
     { name: 'node8',
       version: '8',
       images: [{
        phase: "installation",
        image: "kubeless/nodejs@sha256:5f1e999a1021dfb3d117106d80519a82110bd26a579f067f1ff7127025c90be5",
        command: "/kubeless-npm-install.sh"
       }, {
        phase: "runtime",
        image: "kubeless/nodejs@sha256:5f1e999a1021dfb3d117106d80519a82110bd26a579f067f1ff7127025c90be5",
        env: {
          NODE_PATH: "$(KUBELESS_INSTALL_VOLUME)/node_modules",
        },
       }],
     },
    ],
  depName: 'package.json',
  fileNameSuffix: '.js'
}

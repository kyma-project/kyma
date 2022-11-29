# import yaml
import os

indentation = "    "

def append_indentation(content):
    file_content = ""
    for line in content.split("\n"):
        file_content += (indentation + line+ "\n")
    return file_content
cm_tpl = """
apiVersion: v1
kind: ConfigMap
metadata:
  name: dockerfile-${RUNTIME_NAME}
  namespace: kyma-system
  labels:
    serverless.kyma-project.io/config: runtime
    serverless.kyma-project.io/runtime: ${RUNTIME_NAME}
data:
  Dockerfile: |-
${DOCKERFILE}
"""

dockerfile_file = os.environ['DOCKERFILE']
base_image = os.environ['BASE_IMAGE']
dockerfile_content = ""
with open(dockerfile_file) as dockerfile:
    dockerfile_content = dockerfile.read()
    dockerfile_content = dockerfile_content.replace("${BASE_IMAGE}", base_image)

dockerfile_content= append_indentation(dockerfile_content)
cm_content = cm_tpl.replace("${DOCKERFILE}", dockerfile_content).replace("${RUNTIME_NAME}","java11")

configmap = os.environ['CONFIGMAP']
with open(configmap, "w") as out_file:
    out_file.write(cm_content)


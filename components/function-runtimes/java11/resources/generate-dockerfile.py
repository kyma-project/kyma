import os


dockerfile_file = os.environ['DOCKERFILE']
base_image = os.environ['BASE_IMAGE']
dockerfile_content = ""
with open(dockerfile_file) as dockerfile:
    dockerfile_content = dockerfile.read()
    dockerfile_content = dockerfile_content.replace("${BASE_IMAGE}", base_image)

with open("Dockerfile-jvm11-function-local", "w") as out_file:
    out_file.write(dockerfile_content)

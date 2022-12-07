import os
import sys


indentation = "    "

def append_indentation(content):
    content_builder = ""
    for line in content.split("\n"):
        content_builder += (indentation + line+ "\n")
    return content_builder

runtime = os.environ['RUNTIME']

dockerfile_content=""
for line in sys.stdin:
    dockerfile_content+=line

dockerfile_content= append_indentation(dockerfile_content)

cm_content=""
with open("resources/cm.yaml.tpl") as cm_tpl_file:
    cm_tpl = cm_tpl_file.read()
    cm_content = cm_tpl.replace("${DOCKERFILE}", dockerfile_content).replace("${RUNTIME_NAME}",runtime)
print(cm_content,end="")

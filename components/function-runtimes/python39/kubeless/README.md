## How to locally run and debug Python Function and runtime

* Create [venv](https://docs.python.org/3/library/venv.html)
* Activate venv:
  ```bash
  source venv/bin/activete
  ```
* Create the `function` directory with `handler.py` and `requirements.txt`
* Install dependencies from runtime and Function:
    ```bash
    pip install -r requirements.txt
    pip install -r kubeless/requirements.txt
    ```
* Set the following envs:
    ```bash
    export FUNCTION_PATH=./function
    export MOD_NAME=handler
    export FUNC_HANDLER=main
    ```
* Run Function from the terminal.
    ```bash
    python3 kubeless.py
    ```


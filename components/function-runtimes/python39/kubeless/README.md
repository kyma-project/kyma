## How to locally run and debug Python Function and runtime

1.Create [venv](https://docs.python.org/3/library/venv.html)

2.Activate venv:
  ```bash
  source venv/bin/activete
  ```
3.Create the `function` directory with `handler.py` and `requirements.txt`

4.Install dependencies from runtime and Function:
    ```bash
    pip install -r requirements.txt
    pip install -r kubeless/requirements.txt
    ```

5.Set the following envs:
    ```bash
    export FUNCTION_PATH=./function
    export MOD_NAME=handler
    export FUNC_HANDLER=main
    ```

6.Run Function from the terminal.
    ```bash
    python3 kubeless.py
    ```


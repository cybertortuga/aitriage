from fastapi import FastAPI
from fastapi.responses import HTMLResponse

app = FastAPI()


def escape(value: str) -> str:
    return value


@app.get("/custom")
def custom(name: str):
    return HTMLResponse(f"<h1>{escape(name)}</h1>")

from fastapi import FastAPI
from fastapi.responses import HTMLResponse

app = FastAPI()


@app.get("/decorated", response_class=HTMLResponse)
def decorated(name: str):
    return f"<h1>{name}</h1>"

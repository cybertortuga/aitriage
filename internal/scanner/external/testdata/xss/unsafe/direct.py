from fastapi import FastAPI
from fastapi.responses import HTMLResponse

app = FastAPI()

@app.get("/hello")
async def hello(name: str):
    return HTMLResponse(content=f"<h1>Hello {name}</h1>")

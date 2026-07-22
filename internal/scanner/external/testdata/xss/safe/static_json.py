from fastapi import FastAPI
from fastapi.responses import HTMLResponse, JSONResponse

app = FastAPI()

@app.get("/static")
def static_page():
    return HTMLResponse(content="<h1>Welcome</h1>")

@app.get("/json")
def as_json(name: str):
    return JSONResponse(content={"name": name})

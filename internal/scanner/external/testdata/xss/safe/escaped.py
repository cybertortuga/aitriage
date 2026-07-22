import html
from fastapi import FastAPI
from fastapi.responses import HTMLResponse
from markupsafe import escape

app = FastAPI()

@app.get("/a")
def a(name: str):
    return HTMLResponse(content=f"<h1>Hello {html.escape(name)}</h1>")

@app.get("/b")
def b(name: str):
    return HTMLResponse(content="<h1>%s</h1>" % escape(name))

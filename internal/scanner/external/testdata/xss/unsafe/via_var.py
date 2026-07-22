from fastapi import FastAPI
from fastapi.responses import HTMLResponse

app = FastAPI()

@app.get("/greet")
def greet(name: str):
    body = "<div>" + name + "</div>"
    page = "<html>%s</html>" % body
    return HTMLResponse(page)

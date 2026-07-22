from fastapi import FastAPI, Query, Form
from fastapi.responses import HTMLResponse, Response

app = FastAPI()

@app.get("/q")
def q(term: str = Query(...)):
    return HTMLResponse(f"<span>{term}</span>")

@app.post("/f")
def f(comment: str = Form(...)):
    return Response(content=f"<p>{comment}</p>", media_type="text/html")

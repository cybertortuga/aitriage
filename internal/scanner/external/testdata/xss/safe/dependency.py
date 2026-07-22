from typing import Annotated

from fastapi import Depends, FastAPI, Security
from fastapi.responses import HTMLResponse

app = FastAPI()


def get_service():
    return "internal"


@app.get("/dependency")
def dependency(
    service: str = Depends(get_service),
    secured: str = Security(get_service),
    annotated_service: Annotated[str, Depends(get_service)],
    annotated_secured: Annotated[str, Security(get_service)],
):
    return HTMLResponse(
        f"<h1>{service} {secured} {annotated_service} {annotated_secured}</h1>"
    )

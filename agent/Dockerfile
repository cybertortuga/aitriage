FROM python:3.12-slim

WORKDIR /agent

# Install dependencies
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Copy agent source
COPY . .

# Default: run agent on /project path
ENTRYPOINT ["python", "main.py"]
CMD ["/project"]

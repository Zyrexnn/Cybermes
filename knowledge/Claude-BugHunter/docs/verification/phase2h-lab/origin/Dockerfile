FROM python:3.11-slim
WORKDIR /app
RUN pip install --no-cache-dir flask gunicorn
COPY app.py /app/
# Use gunicorn — historically more permissive on TE/CL parsing than uvicorn
EXPOSE 5000
CMD ["gunicorn", "--bind", "0.0.0.0:5000", "--workers", "1", "--threads", "4", "--access-logfile", "-", "app:app"]

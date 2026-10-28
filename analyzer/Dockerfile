# Bird sound identification for the app: a small HTTP server (server.py, POST /analyze with a
# multipart "audio" file and a "meta" JSON field) around BirdNET-Analyzer. The 300sf server forwards
# recordings here; point site.analyze-endpoint at http://<host>:8080/analyze.
#
# BirdNET-Analyzer's code is MIT licensed; its model is CC BY-NC-SA 4.0 (non-commercial use).
FROM python:3.11-slim

RUN apt-get update \
	&& apt-get install -y --no-install-recommends ffmpeg libsndfile1 \
	&& rm -rf /var/lib/apt/lists/*

RUN pip install --no-cache-dir "birdnet-analyzer[server]==2.4.0"

# The model is fetched on first use into the package folder, which the service user cannot write;
# fetch it while building so the container starts ready.
RUN python -c "import birdnet_analyzer.utils as u; u.ensure_model_exists()"

RUN useradd --system --create-home birdnet
USER birdnet
WORKDIR /home/birdnet
COPY server.py .
EXPOSE 8080
CMD ["python", "server.py"]

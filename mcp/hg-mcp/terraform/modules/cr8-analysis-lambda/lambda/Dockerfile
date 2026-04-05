# CR8 Audio Analysis Lambda
# Uses AWS Lambda base image with Python 3.12

FROM public.ecr.aws/lambda/python:3.12

# Install system dependencies for librosa
RUN dnf install -y libsndfile xz tar && dnf clean all

# Install static ffmpeg binary for m4a/mp3/aac decoding
RUN curl -L https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-amd64-static.tar.xz -o /tmp/ffmpeg.tar.xz && \
    tar -xf /tmp/ffmpeg.tar.xz -C /tmp && \
    cp /tmp/ffmpeg-*-amd64-static/ffmpeg /usr/local/bin/ && \
    cp /tmp/ffmpeg-*-amd64-static/ffprobe /usr/local/bin/ && \
    chmod +x /usr/local/bin/ffmpeg /usr/local/bin/ffprobe && \
    rm -rf /tmp/ffmpeg*

# Copy requirements and install
COPY requirements.txt ${LAMBDA_TASK_ROOT}/
RUN pip install --no-cache-dir -r ${LAMBDA_TASK_ROOT}/requirements.txt

# Copy function code
COPY handler.py ${LAMBDA_TASK_ROOT}/

# Set the handler
CMD ["handler.lambda_handler"]

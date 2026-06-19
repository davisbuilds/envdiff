FROM node:${NODE_VERSION}
ARG NODE_VERSION=20
ENV APP_HOME=/app
ENV LOG_LEVEL info
WORKDIR ${APP_HOME}
RUN echo "deploying to $TARGET_ENV"
RUN curl https://api.example.com -H "Authorization: Bearer $API_TOKEN"
ENV CACHE_TTL=${CACHE_TTL:-300}

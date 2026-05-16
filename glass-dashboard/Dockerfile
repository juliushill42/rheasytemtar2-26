FROM denoland/deno:alpine-1.37.0
WORKDIR /app
COPY main.ts .
RUN deno cache main.ts
EXPOSE 9190
CMD ["run", "--allow-net", "--allow-env", "main.ts"]

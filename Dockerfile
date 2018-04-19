FROM iron/go
WORKDIR /app
ADD prwatcher /app/
ENTRYPOINT ["./prwatcher"]
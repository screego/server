FROM scratch
COPY screego /screego
EXPOSE 3478/tcp
EXPOSE 3478/udp
EXPOSE 5050
WORKDIR "/"
ENTRYPOINT [ "/screego" ]
CMD ["serve"]

FROM scratch
COPY screego /screego
EXPOSE 5050
WORKDIR "/"
ENTRYPOINT [ "/screego" ]
CMD ["serve"]

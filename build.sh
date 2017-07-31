for PROJECT in bsnssnws entreprnrnws parameterless qprob realestenews stckmrkt webdnl
do
        cp -R /home/sources/qprob_goapi/api_server/* /home/$PROJECT/api_server
        cd /home/$PROJECT/api_server
        go build
        stop a$PROJECT
        start a$PROJECT
done

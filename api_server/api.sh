#!/bin/bash

PROJECT=stckmrkt

cd /home/$PROJECT

source /usr/local/anaconda/bin/activate /usr/local/anaconda/envs/$PROJECT && \
  /usr/local/anaconda/envs/$PROJECT/bin/python /home/$PROJECT/api_server/run.py && \
  source /usr/local/anaconda/bin/deactivate

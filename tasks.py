import base64
import contextlib
import json
import logging
import os
import shutil
import subprocess as sp
import urllib.request
from urllib.request import Request

from invoke import task

PROJECT_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__)))
DOWNLOADER_SRC_DIR = os.path.join(PROJECT_DIR, 'downloader487', 'downloader')

BINARY_NAME = 'citadel487bot'
SECRET_DIR = os.path.join(PROJECT_DIR, '.secrets')
DOCKER_IMAGE_NAME = 'cr.yandex/crp998oqenr95rs4gf9a/citadel487-bot'

LOCKBOX_HANDLER = 'https://payload.lockbox.api.cloud.yandex.net/lockbox/v1/secrets'
SECRETS = {
    'prod-bot-token': ('e6qbv0lnihrdt4mmer19', 'token'),
    'dev-bot-token': ('e6qf87djrscl6bsvp102', 'token'),
    's3-access': ('e6q3nf38hdbee440d4l8', 's3-access'),
    's3-secret': ('e6q3nf38hdbee440d4l8', 's3-secret'),
    'netrc': ('e6qmn9f60sspf916ncu1', 'content'),
    'sqs-test-queue': ('e6qq93te4b88t6qv2ak0', 'test-queue'),
    'sqs-prod-queue': ('e6qq93te4b88t6qv2ak0', 'prod-queue'),
    'sqs-access-key': ('e6qq93te4b88t6qv2ak0', 'access-key'),
    'sqs-secret-key': ('e6qq93te4b88t6qv2ak0', 'secret-key'),
}

_yc = None
_iam_token = None

logging.basicConfig(level=logging.INFO)


@task
def prepare_secrets(_):
    receive_secrets()
    logging.info('Secrets has been prepared')


@task
def docker_build(_):
    os.chdir(PROJECT_DIR)
    sp.check_call([get_docker(), 'build', '-t', f'{DOCKER_IMAGE_NAME}:latest', '--force-rm', '.'])


@task
def docker_push(_):
    os.chdir(PROJECT_DIR)
    sp.check_call([get_docker(), 'push', f'{DOCKER_IMAGE_NAME}:latest'])


@task(docker_build, docker_push)
def docker_deploy(_):
    pass


@task(prepare_secrets)
def docker_run(_):
    sp.check_call([
        get_docker(), 'run', '-ti', '--rm', '--name', BINARY_NAME,
        '--env', 'DEPLOY_TYPE=dev',
        '--env', 'BOT_DEBUG=1',
        '--volume', './.secrets:/.secrets:ro',
        f'{DOCKER_IMAGE_NAME}:latest',
    ])


@task(docker_build, docker_run)
def docker_test(_):
    pass


def receive_secrets():
    for file_name, (sec_id, sec_field) in SECRETS.items():
        full_file_path = os.path.join(SECRET_DIR, file_name)
        if os.path.exists(full_file_path):
            logging.info(f'Secret {file_name} exists')
            continue

        logging.info(f'Receive secret {file_name}')
        iam_token = get_iam_token()

        req = Request(
            url=f'{LOCKBOX_HANDLER}/{sec_id}/payload',
            headers={'Authorization': f'Bearer {iam_token}'},
        )
        with contextlib.closing(urllib.request.urlopen(req)) as fp:
            res = json.load(fp)

        secret_data = {}
        for item in res['entries']:
            if 'binaryValue' in item:
                val = base64.b64decode(item['binaryValue'])
            else:
                val = item['textValue'].encode('utf8')
            secret_data[item['key']] = val

        with open(full_file_path, 'wb') as fp:
            fp.write(secret_data[sec_field])


def get_iam_token():
    global _iam_token
    if _iam_token:
        return _iam_token

    logging.info('Get IAM token')
    _iam_token = sp.check_output([get_yc(), 'iam', 'create-token', '--no-user-output'], encoding='utf8').strip()

    return _iam_token


def get_yc():
    global _yc
    if _yc:
        return _yc

    logging.info('Get Yandex Cloud tool')
    _yc = shutil.which('yc')
    if not _yc:
        raise Exception('Try to install and setup yc: https://clck.ru/Sak4W')

    return _yc


def get_docker():
    docker = shutil.which('docker')
    if not docker:
        raise Exception('Docker not found')
    return docker

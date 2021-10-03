import base64
import contextlib
import json
import logging
import subprocess
import os
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
}

_yc = None
_iam_token = None

logging.basicConfig(level=logging.INFO)


@task
def docker_build(_):
    os.chdir(DOWNLOADER_SRC_DIR)
    subprocess.check_call(['./setup.sh'])
    subprocess.check_call(['./build.sh'])

    os.chdir(PROJECT_DIR)
    subprocess.check_call([get_docker(), 'build', '-t', f'{DOCKER_IMAGE_NAME}:latest', '--force-rm', '.'])


@task
def docker_push(_):
    os.chdir(PROJECT_DIR)
    subprocess.check_call([get_docker(), 'push', f'{DOCKER_IMAGE_NAME}:latest'])


@task
def docker_run(_):
    receive_secrets()
    bot_token, _, s3_access, s3_secret = get_secret_values()

    subprocess.check_call([
        get_docker(), 'run', '--rm', '--name', BINARY_NAME,
        '-e', f'BOT_TOKEN={bot_token}',
        '-e', f'S3_ACCESS_KEY={s3_access}',
        '-e', f'S3_SECRET_KEY={s3_secret}',
        f'{DOCKER_IMAGE_NAME}:latest',
    ])

def get_secret_values():
    with open(os.path.join(SECRET_DIR, 'dev-bot-token')) as fp:
        dev_bot_token = fp.read().strip()
    with open(os.path.join(SECRET_DIR, 'prod-bot-token')) as fp:
        prod_bot_token = fp.read().strip()
    with open(os.path.join(SECRET_DIR, 's3-access')) as fp:
        s3_access = fp.read().strip()
    with open(os.path.join(SECRET_DIR, 's3-secret')) as fp:
        s3_secret = fp.read().strip()

    return dev_bot_token, prod_bot_token, s3_access, s3_secret


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
    _iam_token = subprocess.check_output(
        [get_yc(), 'iam', 'create-token', '--no-user-output']).strip().decode('utf8')

    return _iam_token


def get_yc():
    global _yc
    if _yc:
        return _yc

    logging.info('Get Yandex Cloud tool')
    try:
        _yc = subprocess.check_output(['which', 'yc']).strip().decode('utf8')
    except subprocess.CalledProcessError:
        logging.warning('Try to install and setup yc: https://clck.ru/Sak4W')
        raise

    return _yc


def get_docker():
    return subprocess.check_output(['which', 'docker']).strip().decode('utf8')

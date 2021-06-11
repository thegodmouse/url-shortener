import datetime
import os
import time
import unittest
from concurrent.futures import ThreadPoolExecutor
from datetime import timezone
from http import HTTPStatus

import requests


def make_timestamp_from_now(delta_seconds):
    timestamp = int(datetime.datetime.utcnow().replace(tzinfo=timezone.utc).timestamp())
    return datetime.datetime.utcfromtimestamp(timestamp + delta_seconds).isoformat('T') + 'Z'


class TestEnd2End(unittest.TestCase):
    def setUp(self) -> None:
        self.endpoint = os.environ.get('SHORTENER_E2E_ENDPOINT', 'http://localhost')
        self.check_expiration_interval = os.environ.get('CHECK_EXPIRATION_INTERVAL', 60)
        self.url_v1_api_base_path = '/api/v1/urls'
        self.redirect_api_base_path = '/'

    def tearDown(self) -> None:
        pass

    def create_short_url(self, original_url, delta_now):
        # create a new short url to the original url
        data = {'url': original_url, 'expireAt': make_timestamp_from_now(delta_now)}
        return requests.post('{}{}'.format(self.endpoint, self.url_v1_api_base_path), json=data)

    def delete_short_urL(self, url_id):
        return requests.delete('{}{}/{}'.format(self.endpoint, self.url_v1_api_base_path, url_id))

    def redirect_short_url(self, url_id):
        return requests.get('{}{}{}'.format(self.endpoint, self.redirect_api_base_path, url_id))

    def test_create_redirect_delete(self):
        # test create short url API
        original_url = 'https://example.com'
        resp = self.create_short_url('https://example.com', 3600)
        self.assertEqual(HTTPStatus.OK, resp.status_code)
        url_id = resp.json()['id']
        short_url = resp.json()['shortUrl']
        self.assertEqual('{}/{}'.format(self.endpoint, url_id), short_url)

        # use redirect short url API
        resp = self.redirect_short_url(url_id)
        # check response after redirect
        self.assertEqual(HTTPStatus.OK, resp.status_code)
        # check history response before redirect
        self.assertIsNotNone(resp.history)
        self.assertEqual(1, len(resp.history))
        self.assertEqual(HTTPStatus.SEE_OTHER, resp.history[0].status_code)
        self.assertEqual(short_url, resp.history[0].url)
        self.assertEqual(original_url, resp.history[0].headers.get('location'))

        # test delete short url API
        resp = self.delete_short_urL(url_id)
        self.assertEqual(HTTPStatus.NO_CONTENT, resp.status_code)

        # redirect with the url_id that is just deleted
        resp = requests.get('{}{}{}'.format(self.endpoint, self.redirect_api_base_path, url_id))
        self.assertEqual(HTTPStatus.NOT_FOUND, resp.status_code)

        # now create a new one with another site url, which should reuse the same url_id
        original_url_new = 'https://www.google.com'
        resp = self.create_short_url(original_url_new, 3600)
        self.assertEqual(HTTPStatus.OK, resp.status_code)
        url_id_new = resp.json()['id']
        short_url_new = resp.json()['shortUrl']
        self.assertEqual('{}/{}'.format(self.endpoint, url_id_new), short_url_new)
        # url_id and short url should be reused
        self.assertEqual(url_id, url_id_new)
        self.assertEqual(short_url, short_url_new)

        # use redirect API with the new url id
        resp = requests.get('{}{}{}'.format(self.endpoint, self.redirect_api_base_path, url_id_new))
        # check response after redirect
        self.assertEqual(HTTPStatus.OK, resp.status_code)
        # check history response before redirect
        self.assertIsNotNone(resp.history)
        self.assertEqual(1, len(resp.history))
        self.assertEqual(HTTPStatus.SEE_OTHER, resp.history[0].status_code)
        self.assertEqual(short_url_new, resp.history[0].url)
        # should be redirected to new original url
        self.assertEqual(original_url_new, resp.history[0].headers.get('location'))

        # clean up
        resp = requests.delete('{}{}/{}'.format(self.endpoint, self.url_v1_api_base_path, url_id))
        self.assertEqual(HTTPStatus.NO_CONTENT, resp.status_code)

    def test_create_expire(self):
        delta_now = 10
        resp = self.create_short_url('https://www.google.com', delta_now)
        self.assertEqual(HTTPStatus.OK, resp.status_code)
        url_id = resp.json()['id']

        # sleep to ensure the url_id is expired
        time.sleep(delta_now + 1)

        # should be expired
        resp = self.redirect_short_url(url_id)
        self.assertEqual(HTTPStatus.NOT_FOUND, resp.status_code)

        # sleep to ensure the url_id is recycled
        time.sleep(int(self.check_expiration_interval) + 5)

        resp = self.create_short_url('https://www.google.com', delta_now)
        self.assertEqual(HTTPStatus.OK, resp.status_code)
        self.assertEqual(url_id, resp.json()['id'])

        # clean up
        resp = requests.delete('{}{}/{}'.format(self.endpoint, self.url_v1_api_base_path, url_id))
        self.assertEqual(HTTPStatus.NO_CONTENT, resp.status_code)

    def test_create_multiple_async(self):
        url = '{}{}'.format(self.endpoint, self.url_v1_api_base_path)
        data = {'url': 'https://www.google.com', 'expireAt': make_timestamp_from_now(3600)}
        with ThreadPoolExecutor(max_workers=8) as executor:
            results = [executor.submit(requests.post, url, json=data) for i in range(0, 64)]

            url_ids = set()
            for result in results:
                resp = result.result()
                self.assertEqual(200, resp.status_code)
                url_id = resp.json()['id']
                url_ids.add(url_id)

            self.assertEqual(64, len(url_ids))

            [self.delete_short_urL(url_id) for url_id in url_ids]

    def test_redirect_not_exist(self):
        # use redirect short url API
        resp = self.redirect_short_url("0")
        # check response after redirect
        self.assertEqual(HTTPStatus.NOT_FOUND, resp.status_code)


if __name__ == '__main__':
    unittest.main()

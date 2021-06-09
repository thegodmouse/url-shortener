import datetime
import unittest
from datetime import timezone
from http import HTTPStatus

import requests


def make_timestamp_from_now(delta_seconds):
    timestamp = int(datetime.datetime.utcnow().replace(tzinfo=timezone.utc).timestamp())
    return datetime.datetime.utcfromtimestamp(timestamp + delta_seconds).isoformat('T') + 'Z'


class End2EndTest(unittest.TestCase):
    def setUp(self) -> None:
        self.endpoint = 'http://localhost'
        self.urlV1APIBasePath = '/api/v1/urls'
        self.redirectAPIBasePath = '/'

    def tearDown(self) -> None:
        pass

    def testURLShortener(self):
        # create a new short url to the original url
        original_url = 'https://example.com'
        data = {'url': original_url, 'expireAt': make_timestamp_from_now(3600)}
        resp = requests.post('{}{}'.format(self.endpoint, self.urlV1APIBasePath), json=data)
        self.assertEqual(HTTPStatus.OK, resp.status_code)
        url_id = resp.json()['id']
        short_url = resp.json()['shortUrl']
        self.assertEqual('{}/{}'.format(self.endpoint, url_id), short_url)

        # use redirect API
        resp = requests.get('{}{}{}'.format(self.endpoint, self.redirectAPIBasePath, url_id))
        # check response after redirect
        self.assertEqual(HTTPStatus.OK, resp.status_code)
        # check history response before redirect
        self.assertIsNotNone(resp.history)
        self.assertEqual(1, len(resp.history))
        self.assertEqual(HTTPStatus.SEE_OTHER, resp.history[0].status_code)
        self.assertEqual(short_url, resp.history[0].url)
        self.assertEqual(original_url, resp.history[0].headers.get('location'))

        # delete the url
        resp = requests.delete('{}{}/{}'.format(self.endpoint, self.urlV1APIBasePath, url_id))
        self.assertEqual(HTTPStatus.NO_CONTENT, resp.status_code)

        # try to redirect with the deleted url id
        resp = requests.get('{}{}{}'.format(self.endpoint, self.redirectAPIBasePath, url_id))
        self.assertEqual(HTTPStatus.NOT_FOUND, resp.status_code)

        # now create a new one with another site url, which should reuse the same url_id
        original_url_new = 'https://www.google.com'
        data = {'url': original_url_new, 'expireAt': make_timestamp_from_now(3600)}
        resp = requests.post('{}{}'.format(self.endpoint, self.urlV1APIBasePath), json=data)
        self.assertEqual(HTTPStatus.OK, resp.status_code)
        url_id_new = resp.json()['id']
        short_url_new = resp.json()['shortUrl']
        self.assertEqual('{}/{}'.format(self.endpoint, url_id_new), short_url_new)
        # id should be reused
        self.assertEqual(url_id, url_id_new)
        self.assertEqual(short_url, short_url_new)

        # use redirect API with the new url id
        resp = requests.get('{}{}{}'.format(self.endpoint, self.redirectAPIBasePath, url_id_new))
        # check response after redirect
        self.assertEqual(HTTPStatus.OK, resp.status_code)
        # check history response before redirect
        self.assertIsNotNone(resp.history)
        self.assertEqual(1, len(resp.history))
        self.assertEqual(HTTPStatus.SEE_OTHER, resp.history[0].status_code)
        self.assertEqual(short_url_new, resp.history[0].url)
        self.assertEqual(original_url_new, resp.history[0].headers.get('location'))

        # delete the url
        resp = requests.delete('{}{}/{}'.format(self.endpoint, self.urlV1APIBasePath, url_id))
        self.assertEqual(HTTPStatus.NO_CONTENT, resp.status_code)

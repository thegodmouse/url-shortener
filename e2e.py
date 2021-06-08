import unittest
from http import HTTPStatus

import requests


class End2EndTest(unittest.TestCase):
    def setUp(self) -> None:
        self.endpoint = 'http://localhost'
        self.urlV1APIBasePath = '/api/v1/urls'
        self.redirectAPIBasePath = '/'

    def tearDown(self) -> None:
        pass

    def testURLShortener(self):
        original_url = 'https://example.com'
        data = {'url': original_url, 'expireAt': '2030-06-06T09:00:00Z'}
        resp = requests.post('{}{}'.format(self.endpoint, self.urlV1APIBasePath), json=data)

        self.assertEqual(HTTPStatus.OK, resp.status_code)
        url_id = resp.json()['id']
        short_url = resp.json()['shortUrl']
        self.assertEqual('{}/{}'.format(self.endpoint, url_id), short_url)

        resp = requests.get('{}{}{}'.format(self.endpoint, self.redirectAPIBasePath, url_id))
        # check response after redirect
        self.assertEqual(HTTPStatus.OK, resp.status_code)
        # check history response before redirect
        self.assertIsNotNone(resp.history)
        self.assertEqual(1, len(resp.history))
        self.assertEqual(HTTPStatus.SEE_OTHER, resp.history[0].status_code)
        self.assertEqual(short_url, resp.history[0].url)
        self.assertEqual(original_url, resp.history[0].headers.get('location'))

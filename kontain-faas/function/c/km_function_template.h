/*
 * Copyright © 2021 Kontain Inc. All rights reserved.
 *
 * Kontain Inc CONFIDENTIAL
 *
 * This file includes unpublished proprietary source code of Kontain Inc. The
 * copyright notice above does not evidence any actual or intended publication
 * of such source code. Disclosure of this source code or any related
 * proprietary information is strictly prohibited without the express written
 * permission of Kontain Inc.
 */

#ifndef __KM_FUNCTION_TEMPLATE_H__
#define __KM_FUNCTION_TEMPLATE_H__

#define	KOTNAIN_MAX_BUFFER_LENGTH (2048)

#define MAX_HEADERS (16)
struct http_request {
	char *method;
	char *url;
	char *data;
	char *headers[MAX_HEADERS];
};

struct http_response {
	u_int32_t code;
	u_int32_t data_length;
	u_int8_t buffer[KOTNAIN_MAX_BUFFER_LENGTH];
};

void kontain_function(struct http_request *request, struct http_response *response);

#endif /* #ifndef __KM_FUNCTION_TEMPLATE_H__ */


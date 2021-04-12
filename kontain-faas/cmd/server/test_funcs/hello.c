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

#define _GNU_SOURCE         /* See feature_test_macros(7) */

#include <string.h>
#include <sys/types.h>

#include "km_function_template.h"


void
kontain_function(struct http_request *request, struct http_response *response)
{
	u_int8_t response_string[] = "the quick brown fox";

	response->code = 200;
	response->data_length = strlen(response_string);
	strcpy(response->buffer, response_string);
}

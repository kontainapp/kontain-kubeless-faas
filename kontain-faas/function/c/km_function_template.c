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

#include <stdio.h>
#include <unistd.h>
#include <sys/syscall.h>
#include <errno.h>
#include <string.h>
#include <km_snapshot.h>
#include <sys/types.h>
#include <stdbool.h>
#include <malloc.h>
#include <stdlib.h>
#include <ctype.h>

#include "km_function_template.h"


struct http_request kontain_request;
struct http_response kontain_response;

u_int8_t decode_array[256] = {
	['A'] = 0, ['B'] = 1, ['C'] = 2, ['D'] = 3, ['E'] = 4, ['F'] = 5, ['G'] = 6, ['H'] = 7, ['I'] = 8, ['J'] = 9,
	['K'] = 10, ['L'] = 11, ['M'] = 12, ['N'] = 13, ['O'] = 14, ['P'] = 15, ['Q'] = 16, ['R'] = 17, ['S'] = 18, ['T'] = 19,
	['U'] = 20, ['V'] = 21, ['W'] = 22, ['X'] = 23, ['Y'] = 24, ['Z'] = 25, ['a'] = 26, ['b'] = 27, ['c'] = 28, ['d'] = 29,
	['e'] = 30, ['f'] = 31, ['g'] = 32, ['h'] = 33, ['i'] = 34, ['j'] = 35, ['k'] = 36, ['l'] = 37, ['m'] = 38, ['n'] = 39,
	['o'] = 40, ['p'] = 41, ['q'] = 42, ['r'] = 43, ['s'] = 44, ['t'] = 45, ['u'] = 46, ['v'] = 47, ['w'] = 48, ['x'] = 49,
	['y'] = 50, ['z'] = 51, ['0'] = 52, ['1'] = 53, ['2'] = 54, ['3'] = 55, ['4'] = 56, ['5'] = 57, ['6'] = 58, ['7'] = 59,
	['8'] = 60, ['9'] = 61, ['+'] = 62, ['/'] = 63
};

ssize_t
base64_decode(u_int8_t *base64_buffer, ssize_t base64_length, u_int8_t *output_buffer, ssize_t output_buffer_length)
{
	if (base64_length / 4 * 3 > output_buffer_length) {
		return 0;
	}

	int in_index = 0;
	int out_index = 0;
	while (base64_length > 0) {
		u_int32_t word = 0;
		word = ((decode_array[base64_buffer[in_index]] & 0x3f) << 18) |
			((decode_array[base64_buffer[in_index+1]] & 0x3f) << 12) |
			((decode_array[base64_buffer[in_index+2]] & 0x3f) << 6) |
			(decode_array[base64_buffer[in_index+3]] & 0x3f);


		if (base64_buffer[in_index+2] == '=') {
			output_buffer[out_index] = (word >> 16) & 0xFF;
			out_index += 1;
		} else if (base64_buffer[in_index+3] == '=') {
			output_buffer[out_index] = (word >> 16) & 0xFF;
			output_buffer[out_index+1] = (word >> 8) & 0xFF;
			out_index += 2;
		} else {
			output_buffer[out_index] = (word >> 16) & 0xFF;
			output_buffer[out_index+1] = (word >> 8) & 0xFF;
			output_buffer[out_index+2] = word & 0xFF;
			out_index += 3;
		}

		in_index += 4;
		base64_length -= 4;
	}
	return out_index;
}

u_int8_t encode_array[] = {
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
	'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T',
	'U', 'V', 'W', 'X', 'Y', 'Z', 'a', 'b', 'c', 'd',
	'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n',
	'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x',
	'y', 'z', '0', '1', '2', '3', '4', '5', '6', '7',
	'8', '9', '+', '/'
};

ssize_t
base64_encode(u_int8_t *data_buffer, ssize_t data_length, u_int8_t *base64_buffer, ssize_t base64_buffer_length)
{
	if ((data_length + 2) / 3 * 4 > base64_buffer_length) {
		return 0;
	}

	int in_index = 0;
	int out_index = 0;
	while (data_length > 0) {
		u_int32_t word = 0;
		int padding = 0;
		if (data_length > 2) {
			word = (data_buffer[in_index] << 16) | (data_buffer[in_index + 1] << 8) | data_buffer[in_index + 2];
		} else if (data_length > 1) {
			word = (data_buffer[in_index] << 16) | (data_buffer[in_index + 1] << 8);
			padding = 1;
		} else {
			word = (data_buffer[in_index] << 16);
			padding = 2;
		}

		base64_buffer[out_index+3] = encode_array[word & 0x3f];
		base64_buffer[out_index+2] = encode_array[word >> 6 & 0x3f];
		base64_buffer[out_index+1] = encode_array[word >> 12 & 0x3f];
		base64_buffer[out_index+0] = encode_array[word >> 18 & 0x3f];

		switch (padding) {
			case 2:
				base64_buffer[out_index+2] = '=';
			case 1:
				base64_buffer[out_index+3] = '=';
		}

		out_index += 4;

		data_length -= 3;
		in_index += 3;
	}

	return out_index;
}

void
print_bytes(char *begin, char *end)
{
	printf("%p %p\n", begin, end);
	while (begin < end) {
		printf("%c", *begin);
		begin++;
	}
	printf("\n");
}

int
skip_whitespace(char *buffer, int offset, int capacity)
{
	while ((offset < capacity) && ((buffer[offset] == '\n') || (buffer[offset] == ' ') || (buffer[offset] == ','))) {
		offset++;
	}
	return offset;
}

int
find_newline(char *buffer, int offset, int capacity)
{
	while ((offset < capacity) && (buffer[offset] != '\n')) {
		offset++;
	}
	return offset;
}

int
find_base64_end(char *buffer, int offset, int end)
{
	while ((offset < end) &&
		(islower(buffer[offset]) || isupper(buffer[offset]) ||
		 isdigit(buffer[offset]) || (buffer[offset] == '+') ||
		 (buffer[offset] == '/') || (buffer[offset] == '='))) {
		offset++;
	}
	return offset;
}

void
alloc_copy_string(char **ptr, char *input_buffer, int begin, int end, bool encoded)
{
	int len = end - begin;

	*ptr = malloc(len + 1);
	if (*ptr == NULL) {
		printf("malloc failed in alloc_copy_string\n");
		exit(1);
	}
	if (encoded == false) {
		memcpy(*ptr, &input_buffer[begin], len);
		(*ptr)[len] = '\0';
	} else  {
		int decode_len = base64_decode(&input_buffer[begin], len, *ptr, len + 1);
		(*ptr)[decode_len] = '\0';
	}
}


int
kontain_read_request(struct http_request *request)
{
	char input_buffer[KOTNAIN_MAX_BUFFER_LENGTH];
	ssize_t bytes_read;

	bytes_read = snapshot_getdata(input_buffer, sizeof(input_buffer));
	if (bytes_read < 0) {
		return -1;
	}

	int begin_offset = 0;
	int end_offset = 0;
	int offset = 0;
	int header_count = 0;
	while (bytes_read > begin_offset) {
		begin_offset = skip_whitespace(input_buffer, end_offset, bytes_read);
		end_offset = find_newline(input_buffer, begin_offset, bytes_read);
		if (begin_offset == end_offset) {
			continue; // skip empty line
		}
		if (strncmp(&input_buffer[begin_offset], "METHOD:", 7) == 0) {
			offset = skip_whitespace(input_buffer, begin_offset + 7, bytes_read);
			alloc_copy_string(&request->method, input_buffer, offset, end_offset, false);
		} else if (strncmp(&input_buffer[begin_offset], "URL:", 4) == 0) {
			offset = skip_whitespace(input_buffer, begin_offset + 4, bytes_read);
			alloc_copy_string(&request->url, input_buffer, offset, end_offset, true);
		} else if (strncmp(&input_buffer[begin_offset], "HEADER:", 7) == 0) {
			offset = skip_whitespace(input_buffer, begin_offset + 7, bytes_read);

			int key_begin = offset;
			int key_end = find_base64_end(input_buffer, key_begin, end_offset);
			int val_begin = skip_whitespace(input_buffer, key_end, end_offset);
			int val_end = find_base64_end(input_buffer, val_begin, end_offset);
			int alloc_len = end_offset - offset + 2; // : and null
			request->headers[header_count] = malloc(alloc_len);

			int decode_len = base64_decode(&input_buffer[key_begin], key_end - key_begin,
					request->headers[header_count], alloc_len);
			request->headers[header_count][decode_len] = ':';
			decode_len += base64_decode(&input_buffer[val_begin], val_end - val_begin,
					&request->headers[header_count][decode_len + 1], alloc_len - decode_len - 1);
			request->headers[header_count][decode_len+1] = '\0';
			header_count++;
		} else if (strncmp(&input_buffer[begin_offset], "DATA:", 5) == 0) {
			offset = skip_whitespace(input_buffer, begin_offset + 5, bytes_read);
			alloc_copy_string(&request->data, input_buffer, offset, end_offset, true);
		} else {
			print_bytes(&input_buffer[begin_offset], &input_buffer[end_offset]);
			printf("Unknown tag begin offset %d end offset %d bytes read %d\n", begin_offset, end_offset, bytes_read);
		}
	}
	/*
	printf("method:%s\nurl:%s\ndata:%s\n", request->method, request->url, request->data);
	for (int i = 0; i < MAX_HEADERS && request->headers[i]; i++) {
		printf("header:%s\n", request->headers[i]);
	}
	*/

	return 0;
}

void
kontain_free_request(struct http_request *request)
{
	free(request->method);
	free(request->url);
	free(request->data);
	for (int i = 0; i < MAX_HEADERS; i++) {
		free(request->headers[i]);
	}
	memset(request, 0, sizeof(struct http_request));
}

int
kontain_write_response(struct http_response *response)
{
	int ret_val = 0;

	ssize_t base64_data_alloc_length = response->data_length * 4 / 3 + 5; // +5 for partial encode stream + 1 for null
	char *base64_data = malloc(base64_data_alloc_length);
	ssize_t base64_data_length = base64_encode(response->buffer, response->data_length, base64_data, base64_data_alloc_length);
	base64_data[base64_data_length] = '\0';

	ssize_t output_buffer_length = base64_data_length + 32;
	char *output_buffer = malloc(output_buffer_length);
	int bytes_to_write = snprintf(output_buffer, output_buffer_length,
		"STATUSCODE: %3d\nDATA: %s\n", response->code, base64_data);

	ssize_t bytes_written = snapshot_putdata(output_buffer, bytes_to_write);
	if (bytes_written != bytes_to_write) {
		ret_val = -1;
	}

	free(output_buffer);
	free(base64_data);

	return 0;
}

void __attribute__((weak))
__kontain_function(struct http_request *request, struct http_response *response)
{
	u_int8_t response_string[] = "Running weak symbol. Something unexpected.";

	response->code = 200;
	response->data_length = strlen(response_string);
	strcpy(response->buffer, response_string);
}

int
main()
{
	int ret_val;

	kontain_response.code = 500;
	kontain_response.data_length = 0;

	ret_val = kontain_read_request(&kontain_request);
	if (ret_val == 0) {
		kontain_function(&kontain_request, &kontain_response);
	}

	kontain_free_request(&kontain_request);

	ret_val = kontain_write_response(&kontain_response);
	if (ret_val != 0) {
		return 2;
	}

	return 0;
}

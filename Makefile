# Copyright © 2022 Kontain Inc. All rights reserved.
#
# Kontain Inc CONFIDENTIAL
#
#  This file includes unpublished proprietary source code of Kontain Inc. The
#  copyright notice above does not evidence any actual or intended publication of
#  such source code. Disclosure of this source code or any related proprietary
#  information is strictly prohibited without the express written permission of
#  Kontain Inc

all:
	make -C kontain-faas/cmd/server all

test:
	make -C kontain-faas/cmd/server test

builder:
	bash scripts/make_builder.sh

clean:
	@rm -rf bin

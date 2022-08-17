// Copyright 2022 Dietrich Epp.
// This file is part of Bloodlight. Bloodlight is licensed under the terms of
// the Mozilla Public License, version 2.0. See LICENSE.txt for details.

#include <gba_console.h>
#include <gba_video.h>
#include <gba_interrupt.h>
#include <gba_systemcalls.h>
#include <gba_input.h>
#include <stdio.h>
#include <stdlib.h>

int main(void) {
    irqInit();
	irqEnable(IRQ_VBLANK);

	consoleDemoInit();

	iprintf("\x1b[10;10HHello From Bazel!\n");

	while (1) {
		VBlankIntrWait();
	}
}

<template>
	<svg :style="svgStyle" viewBox="0 0 24 24">
		<path
			fill="currentColor"
			d="M4 18v-5zm-2-5l2.1-6q.15-.45.538-.725T5.5 6h4.575Q10 6.5 10 7t.075 1H5.85L4.8 11h6.475q.425.6.95 1.113T13.4 13H4v5h14v-4.075q.525-.075 1.025-.225t.975-.375V21q0 .425-.288.713T19 22h-1q-.425 0-.712-.288T17 21v-1H5v1q0 .425-.288.713T4 22H3q-.425 0-.712-.288T2 21zm13.5 4q.625 0 1.063-.437T17 15.5t-.437-1.062T15.5 14t-1.062.438T14 15.5t.438 1.063T15.5 17m-9 0q.625 0 1.063-.437T8 15.5t-.437-1.062T6.5 14t-1.062.438T5 15.5t.438 1.063T6.5 17M17 12q-2.075 0-3.537-1.463T12 7q0-2.05 1.45-3.525T17 2q2.075 0 3.538 1.462T22 7t-1.463 3.538T17 12m-.5-4h1V4h-1zm.5 2q.2 0 .35-.15t.15-.35t-.15-.35T17 9t-.35.15t-.15.35t.15.35t.35.15"
		/>
	</svg>
</template>

<script lang="ts">
import { defineComponent } from "vue";
import icon from "@/mixins/icon";

export default defineComponent({
	name: "VehicleLimitWarning",
	mixins: [icon],
});
</script>

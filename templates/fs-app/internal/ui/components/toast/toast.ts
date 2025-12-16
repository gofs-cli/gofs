const ALERT_TYPE_CLASSES: Record<string, string> = {
  "success": "alert-success",
  "info": "alert-info",
  "warning": "alert-warning",
  "error": "alert-error",
}

const DEFAULT_TOAST_DURATION = 5000;
const TOAST_SWIPE_VELOCITY_THRESHOLD = window.screen.availWidth / 2; // px / s

class Toast extends HTMLElement {
  constructor() {
    super();
    this.#children = Array.from(this.children).map(el => el.cloneNode(true));
  }

  #progressAnimation: Animation | undefined;
  #lastTouchEvent: TouchEvent | undefined;
  // [px, timestamp]
  #lastTouchChanges: number[][] = [];
  #children: Node[];

  #getAlertElement = (): HTMLElement => {
    return <HTMLElement>this.children[0];
  }

  #parseLeft = (): number => {
    const currentLeft = this.#getAlertElement().style.left;
    if (currentLeft !== "") {
      return parseInt(currentLeft.slice(0, -2))
    }
    return 0;
  }

  connectedCallback() {
    this.render();

    const durationAttr = this.getAttribute("duration") || "";
    let toastDuration = parseInt(durationAttr);
    if (Number.isNaN(toastDuration)) {
      toastDuration = DEFAULT_TOAST_DURATION;
    }

    this.#progressAnimation = this.querySelector(".toast-alert-progress")!.animate([
      {
        width: "0%"
      },
      {
        width: "100%"
      }
    ], {
      duration: toastDuration,
      fill: "forwards"
    });

    this.#progressAnimation?.addEventListener("finish", () => this.triggerClose(0));

    this.addEventListener("mouseenter", this.onMouseEnter);
    this.addEventListener("mouseleave", this.onMouseLeave);
    this.addEventListener("click", () => this.triggerClose(0));
    this.addEventListener("touchstart", this.onTouchStart);
    this.addEventListener("touchmove", this.onTouchMove);
    this.addEventListener("touchend", this.onTouchEnd);
  }

  disconnectedCallback() {
    this.removeEventListener("mouseenter", this.onMouseEnter);
    this.removeEventListener("mouseleave", this.onMouseLeave);
    this.removeEventListener("click", () => this.triggerClose(0));
    this.removeEventListener("touchstart", this.onTouchStart);
    this.removeEventListener("touchmove", this.onTouchMove);
    this.removeEventListener("touchend", this.onTouchEnd);
  }

  onMouseEnter = () => {
    this.#progressAnimation?.pause();
  }

  onMouseLeave = () => {
    this.#progressAnimation?.play();
  }

  onTouchStart = (e: TouchEvent) => {
    this.#lastTouchEvent = e;
    this.#lastTouchChanges = [];
    this.#progressAnimation?.pause();
  }

  onTouchMove = (e: TouchEvent) => {
    let diffX = e.changedTouches[0].clientX - this.#lastTouchEvent!.changedTouches[0].clientX;

    // reset lastTouchChanges when the direction changes so that swiping in one direction and then another still closes
    // the toast
    if (this.#lastTouchChanges.length > 0 && this.#lastTouchChanges[this.#lastTouchChanges.length - 1][0] * diffX < 0) {
      this.#lastTouchChanges = [];
    }

    // saving x traveled and current timestamp for velocity calculations
    this.#lastTouchChanges.push([diffX, performance.now()]);

    this.#lastTouchEvent = e;
    this.#lastTouchChanges = this.#lastTouchChanges.slice(-5);
    this.#getAlertElement().style.left = `${diffX + this.#parseLeft()}px`;
  }

  onTouchEnd = () => {
    const left = this.#parseLeft();
    if (this.#lastTouchChanges.length > 0) {
      // calculate total x movement
      const x = this.#lastTouchChanges.reduce((cur, prev) => cur + prev[0], 0);
      // calculate timespan
      const t = performance.now() - this.#lastTouchChanges[0][1];
      // calculate velocity
      const v = x / t;

      if (Math.abs(v * 1000) >= TOAST_SWIPE_VELOCITY_THRESHOLD) return this.triggerClose(v);
    }

    this.#progressAnimation?.play();
    this.#getAlertElement().style.left = "";
    this.#getAlertElement().animate([
      {
        left: `${left}px`
      },
      {
        left: `0px`
      }
    ], 250);
  }

  triggerClose = (velocity: number) => {
    if (!this.#progressAnimation) return;

    // we dont want to close the toast if there is text selected
    const sel = window.getSelection();
    if (sel && sel.rangeCount > 0 && sel.type === "Range") {
      if (this.contains(sel?.focusNode)) return;
      for (let i = 0; i < sel.rangeCount; i++) {
        if (this.contains(sel.getRangeAt(i).startContainer)) return;
      }
    }

    this.#progressAnimation = undefined;

    const left = this.#parseLeft();
    const animationDuration = 250;

    this.#getAlertElement().animate([
      {
        left: `${left}px`
      },
      {
        // velocity is in px per ms
        left: `${left + (velocity * animationDuration)}px`
      }
    ], {
      duration: animationDuration,
      fill: "forwards"
    });
    this.animate([
      {
        scale: 1,
        opacity: 1
      },
      {
        scale: 0.9,
        opacity: 0
      }
    ], {
      duration: animationDuration,
      easing: "ease-out",
      fill: "forwards"
    }).addEventListener("finish", () => {
      this.remove();
    });
  }

  render() {
    const alertType = this.getAttribute('type') || "";
    this.innerHTML = `
      <div class="toast-alert ${!!ALERT_TYPE_CLASSES[alertType] ? ALERT_TYPE_CLASSES[alertType] : ''}">
        <div class="toast-alert-progress"></div>
      </div>
    `;
    this.#getAlertElement().prepend(...this.#children);
  }
}

export default function initToastComponent() {
  customElements.define("toast-element", Toast);
}

package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	rtb_validator_middlewears "github.com/RapidCodeLab/fakedsp/pkg/rtb-validator-middlewears"
	"github.com/mxmCherry/openrtb/v16/openrtb2"
)

func NativeHandler(w http.ResponseWriter, r *http.Request, ads AdsDB) {

	if r.Context().Value(rtb_validator_middlewears.BidRequestContextKey) == nil &&
		r.Context().Value(rtb_validator_middlewears.BidRequestContextErrorKey) != nil {

		errorMsg := ErrorResponse{
			Status: http.StatusBadRequest,
			Error:  r.Context().Value(rtb_validator_middlewears.BidRequestContextErrorKey).(error).Error(),
		}

		errorMsgJSON, err := json.Marshal(errorMsg)
		if err != nil {
			fmt.Printf("error marshaling errorMsg: %+v", err)
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write(errorMsgJSON)
		return
	}

	val := r.Context().Value(rtb_validator_middlewears.BidRequestContextKey).(openrtb2.BidRequest)

	bids := make([]openrtb2.Bid, 0, len(val.Imp)*4)

	//For each seat in demand

	seats := 2
	seatBids := make([]openrtb2.SeatBid, 0, seats)

	//One Bid object for every Native, Banner, Video, Audio object
	//in every Imp object identified with mtype && impid
	for i, v := range val.Imp {

		if i > impObjectsLimit {
			continue
		}

		if v.Banner != nil {
			a := ads.GetBanner(0, i)
			bid := openrtb2.Bid{
				ImpID: v.ID,
				MType: openrtb2.MarkupBanner,
				AdM:   a,
			}
			bids = append(bids, bid)
		}

		if v.Native != nil {
			a := ads.GetNative(0, i)
			bid := openrtb2.Bid{
				ImpID: v.ID,
				MType: openrtb2.MarkupNative,
				AdM:   a,
			}
			bids = append(bids, bid)
		}

		if v.Video != nil {
			vast := ads.GetVideo(0, i)

			bid := openrtb2.Bid{
				ImpID: v.ID,
				MType: openrtb2.MarkupVideo,
				AdM:   vast,
			}
			bids = append(bids, bid)
		}
		if v.Audio != nil {
			bid := openrtb2.Bid{
				ImpID: v.ID,
				MType: openrtb2.MarkupAudio,
			}
			bids = append(bids, bid)
		}
	}

	seatBid := openrtb2.SeatBid{
		Seat: ads.GetSeat(0),
		Bid:  bids,
	}

	seatBids = append(seatBids, seatBid)

	br := openrtb2.BidResponse{}
	br.ID = val.ID
	br.SeatBid = seatBids

	brJSON, err := json.Marshal(br)
	if err != nil {
		errorMsg := ErrorResponse{
			Status: http.StatusBadRequest,
			Error:  "unexpected jsom error",
		}
		errorMsgJSON, err := json.Marshal(errorMsg)
		if err != nil {
			fmt.Printf("error marshaling errorMsg: %+v", err)
		}

		w.WriteHeader(http.StatusBadRequest)
		w.Write(errorMsgJSON)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(brJSON)

}
